package utils

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/golang/glog"
)

type Conf struct {
	MysqlUser            string   `yaml:"mysql_user"`//yaml：yaml格式 enabled：属性的为enabled
	MysqlPassword    string `yaml:"mysql_password"`
	MysqlHost            string `yaml:"mysql_host"`
	MysqlPort             string `yaml:"mysql_port"`

	RedisMasterName       string   `yaml:"redis_mastername"`//yaml：yaml格式 enabled：属性的为enabled
	RedisPassword         string   `yaml:"redis_password"`
	RedisSentinelAddrs    string   `yaml:"redis_sentineladdrs"`
	RedisDB               int      `yaml:"redis_db"`

	//JaegerAgentUDPAddr    string    `yaml:"jaeger_agent_udp_addr"`
	//JaegerSamplingServerURL  string  `yaml:"jaeger_sampling_server_url"`
}

func (c *Conf) GetConf() (*Conf, error) {

yamlFile, err := ioutil.ReadFile("conf.yaml")
if err != nil {
	glog.Error(err)
	glog.Errorf("yamlFile.Get err   #%v ", err)
	return  nil, err
}
err = yaml.Unmarshal(yamlFile, c)
if err != nil {
	glog.Error(err)
	glog.Fatalf("Unmarshal: %v", err)
	return  nil, err
}

return c, nil
}






