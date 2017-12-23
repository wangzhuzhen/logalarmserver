package utils

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/golang/glog"
)

type Conf struct {
	User  string   `yaml:"user"`//yaml：yaml格式 enabled：属性的为enabled
	Password    string `yaml:"password"`
	MysqlHost    string `yaml:"mysql_host"`
	MysqlPort    string `yaml:"mysql_port"`
}

func (c *Conf) GetConf() (*Conf, error) {

	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		glog.Errorf("yamlFile.Get err   #%v ", err)
		return  nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		glog.Fatalf("Unmarshal: %v", err)
		return  nil, err
	}

	return c, nil
}
