package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"net/url"
	"robfig/cron"
	"strconv"
	"strings"
	"time"
	"yan.site/rep-auto/config"
)

const (
	url_report = "http://xsc.swust.edu.cn/SPCP/Web/Temperature/StuTemperatureInfo"
)

type TemperAndTime struct {
	TimeNowHour   string
	TimeNowMinute string
	Temper1       string
	Temper2       string
	ReSubmiteFlag string
}

func NewTemperAndTime() *TemperAndTime {
	return &TemperAndTime{
		TimeNowHour:   strconv.Itoa(time.Now().Hour()),
		TimeNowMinute: strconv.Itoa(time.Now().Minute()),
		Temper1:       "36",
		Temper2:       RandomAInt(),
	}
}

func main() {

	users := config.GetConf().Users
	log.Println("Report task started")
	log.Println("当前配置上报", len(users), "人")

	c := cron.New()
	c.AddFunc("2 9,13,17,21,22 * * *", func() {
		for _, u := range users {
			go reportTask(u)
		}
	})
	c.Start()
	select {}
}

func reportTask(user config.User) {

	resMsg := ""
	reCount := 5
	for index := 1; index <= reCount; index++ {
		success, msg, temperAndTime := reportTemper(user)
		resMsg = resMsg + "第" + strconv.Itoa(index) + "次尝试: " + msg + "\n"
		if strings.Contains(msg, "0") {
			sendEMail(user, temperAndTime, success, resMsg)
			return
		} else {
			time.Sleep(time.Duration(time.Now().Hour()/4) * time.Minute)
		}
	}
	sendEMail(user, NewTemperAndTime(), false, resMsg)
}

func reportTemper(user config.User) (bool, string, *TemperAndTime) {

	temperAndTime := NewTemperAndTime()

	data := make(url.Values)
	data["TimeNowHour"] = []string{temperAndTime.TimeNowHour}
	data["TimeNowMinute"] = []string{temperAndTime.TimeNowMinute}
	data["Temper1"] = []string{temperAndTime.Temper1}
	data["Temper2"] = []string{temperAndTime.Temper2}
	data["ReSubmiteFlag"] = []string{RandomOrderId()}

	req, err := http.NewRequest("POST", url_report, strings.NewReader(data.Encode()))
	if err != nil {
		msg := "code: -1,创建请求发生错误"
		log.Println(user.Name, "的请求构造：", msg, err)
		return false, msg, temperAndTime
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
	req.Header.Set("Cookie", user.Cookie)

	httpClient := http.DefaultClient
	httpClient.Timeout = 1 * time.Minute
	resp, err := httpClient.Do(req)
	if err != nil {
		msg := "code: -1,站点无响应"
		log.Println(user.Name, "的请求：", msg, err)
		return false, msg, temperAndTime
	}
	if resp.StatusCode >= 300 {
		msg := "code: -1,站点状态码不正确"
		log.Println(user.Name, "的请求：", msg, resp.StatusCode)
		return false, msg, temperAndTime
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := "code: -1,站点超时或返回文本异常"
		log.Println(user.Name, "的请求：", msg, err)
		return false, msg, temperAndTime
	}

	text := string(respBody)
	success := false
	msg := "code: -1,未知异常或错误"
	if strings.Contains(text, "成功") {
		msg = "code: 0,填报成功"
		success = true
	} else if strings.Contains(text, "次数已完成") {
		msg = "code: 0,今日填报次数已完成"
	} else if strings.Contains(text, "小时") {
		msg = "code: 0,当前填报未超过4小时"
	}
	log.Println(user.Name, "填报结果：", msg)

	return success, msg, temperAndTime
}

func sendEMail(user config.User, temperAndTime *TemperAndTime, success bool, reMsg string) error {
	receviceAddr := user.Email
	senderAddr := "1316952971@qq.com"
	authCode := "leqkyvxagbbnhafj"
	host := "smtp.qq.com"
	auth := smtp.PlainAuth("", senderAddr, authCode, host)
	to := []string{receviceAddr}
	nickname := "言言健康"
	subject := "体温自动上报通知-言言技术"
	contentType := "Content-Type:text/plain;charset=UTF-8\r\n"
	body := "name：" + user.Name +
		"\nstudent id: " + user.StudentId +
		"\nreport time: " + temperAndTime.TimeNowHour + ":" + temperAndTime.TimeNowMinute +
		"\nreport temper: " + temperAndTime.Temper1 + "." + temperAndTime.Temper2 + " 摄氏度" +
		"\nreport success：" + strconv.FormatBool(success) +
		"\nreport result：" + reMsg +
		"version：V5" +
		"\n如果上报状态失败，则可能是晚上9点尝试第四次上报，不用在乎。若是下午5点则可能当天由于学校系统重复上报，若上午9点或下午1点上报失败请联系我～"
	msg := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + nickname +
		"<" + senderAddr + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	err := smtp.SendMail(host+":25", auth, senderAddr, to, msg)
	if err != nil {
		err = fmt.Errorf("邮件发送失败%v", err)
		log.Println(err)
		return err
	}
	log.Println("邮件发送成功")
	return nil
}

func RandomOrderId() string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 32; i++ {
		if i == 8 || i == 12 || i == 16 || i == 20 {
			result = append(result, '-')
		}
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func RandomAInt() string {
	randInt := rand.Intn(10)
	return strconv.Itoa(randInt)
}
