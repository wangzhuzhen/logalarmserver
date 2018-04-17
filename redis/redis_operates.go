package redis

import (
	"github.com/go-redis/redis"
	"github.com/golang/glog"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/utils"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/types"
	"encoding/json"
	"strings"
	"strconv"
)

func GetPhoneAndEmailInfo(groupId int) ([]types.Contact, error) {

	var c utils.Conf
	var err error
	_, err=c.GetConf()
	if err !=nil {
		return nil, err
	}

	client := redis.NewFailoverClient(&redis.FailoverOptions{
		//MasterName: "TestMaster",
		MasterName: c.RedisMasterName,
		//Password: "0234kz9*l",
		Password: c.RedisPassword,
		//DB: 9,
		DB: c.RedisDB,
		//SentinelAddrs: []string{"172.25.3.195:26371","172.25.3.195:26372","172.25.3.195:26373"},

		SentinelAddrs: strings.Fields(strings.Replace(c.RedisSentinelAddrs, ",", " ", -1)),
	})

	key := "a_e_g_i_" + strconv.Itoa(groupId) //最后的数字表示报警组Id

	val, err := Get(client, key).Result()
	if err != nil {
		glog.Error(err)
		glog.Errorf("Can not get alarmgroup info ")
		return nil, err
	}


	var pms types.PhonesAndEmails
	if err := json.Unmarshal([]byte(val), &pms); err != nil {
		glog.Error(err)
		return nil, err
	}
		return pms.Contacts, nil
}


func Get(client *redis.Client, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd("get", key)
	client.Process(cmd)
	return cmd
}


