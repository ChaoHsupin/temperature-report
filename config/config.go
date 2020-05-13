package config

import (
	"github.com/winjeg/go-commons/conf"
	"sync"
)

const configFile = "config.yml"

// 配置文件结构
type Config struct {
	Users []User `yaml:"users"`
}

type User struct {
	Name      string `yaml:"name"`
	Email     string `yaml:"email"`
	StudentId string `yaml:"student_id"`
	Passwd    string `yaml:"passwd"`
	Cookie    string `yaml:"cookie"`
}

// 全局变量
var (
	once      sync.Once
	configure *Config
)

func GetConf() *Config {
	if configure != nil {
		return configure
	} else {
		once.Do(getConf)
	}
	return configure
}

func getConf() {
	configure = new(Config)
	err := conf.Yaml2Object(configFile, &configure)
	if err != nil {
		panic(err)
	}
}
