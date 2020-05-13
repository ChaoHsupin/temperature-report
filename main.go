package main

import (
	"fmt"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"
	"yan.site/rep-auto/config"
)

const (
	url_home   = "http://xsc.swust.edu.cn/SPCP/Web/"
	url_report = "http://xsc.swust.edu.cn/SPCP/Web/Temperature/StuTemperatureInfo"
)

func main() {
	log.Println("Task started")

	c := cron.New()
	c.AddFunc("2 9,13,17,21 * * *", func() {

		for _, v := range config.GetConf().Users {
			result := reportTemper(v, 3)

			if result == true {
				log.Println(v.Name + " 填报成功")

			} else {
				log.Println(v.Name + " 填报失败")
			}
			sendEMail(v, result)
		}
	})

	c.Start()
	select {}
}

func reportTemper(user config.User, recount int8) bool {

	if recount == 0 {
		return false
	}

	data := make(url.Values)
	data["ReportTime"] = []string{CurrentTime()}
	data["Temperature"] = []string{"36." + RandomAInt()}
	data["OrderId"] = []string{RandomOrderId()}

	req, err := http.NewRequest("POST", url_report, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println(user.Name+"	create req failed", err)
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
	req.Header.Set("Cookie", user.Cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode >= 300 {
		log.Println(user.Name+"	Request failed", "code: ", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(user.Name+"	Parse text failed", err)
		return false
	}
	println(string(respBody))
	success := strings.Contains(string(respBody), "成功")

	if !success {
		time.Sleep(time.Minute*time.Duration(time.Now().Hour()/3))
		return reportTemper(user, recount-1)
	} else {
		return success
	}
}

func sendEMail(user config.User, suc bool) error {
	receviceAddr := user.Email
	senderAddr := "1316952971@qq.com"
	authCode := "leqkyvxagbbnhafj"
	auth := smtp.PlainAuth("", senderAddr, authCode, "smtp.qq.com")
	to := []string{receviceAddr}
	nickname := "体温健康"
	subject := "体温自动上报通知-言言技术"
	contentType := "Content-Type:text/plain;charset=UTF-8\r\n"
	body := "name：" + user.Name +
		"\nstudent id: "+user.StudentId+
		"\nreport success：" + strconv.FormatBool(suc)+
		"\n如果上报状态失败，则可能是晚上9点尝试第四次上报，不用在乎。若是下午5点则可能当天由于学校系统重复上报，若上午9点或下午1点上报失败请联系我～"
	msg := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + nickname +
		"<" + senderAddr + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	err := smtp.SendMail("smtp.qq.com:25", auth, senderAddr, to, msg)
	if err != nil {
		err = fmt.Errorf("邮件发送失败%v", err)
		log.Println(err)
		return err
	}
	log.Println("邮件发送成功")
	return nil
}

func CurrentTime() string {
	currentTime := time.Now()
	return strconv.Itoa(currentTime.Hour()) + ":" + strconv.Itoa(currentTime.Minute())
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
	return string(rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10))
}
