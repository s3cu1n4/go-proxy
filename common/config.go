package common

import (
	"go-proxy/common/logs"
	"os"

	"github.com/spf13/viper"
)

var Conf = &Config{}

type Config struct {
	Server Server `mapstructure:"server"`
}

type Server struct {
	ServerHost        string `mapstructure:"serverhost"`
	ServerPort        string `mapstructure:"serverport"`
	ServerControlPort string `mapstructure:"controlport"`
	ServerHandlerKey  string `mapstructure:"serverhandlerkey"`
	LocalPort         string `mapstructure:"localport"`
}

func GetConfig(filename string) {
	Conf = &Config{}
	viper.SetConfigType("yaml")
	viper.SetConfigFile(filename)
	err := viper.ReadInConfig()
	if err != nil {
		logs.Error("read config error, user default conf")
		// AddDefaults(Conf)
		Conf.Server.ServerControlPort = "28009"
		Conf.Server.LocalPort = "3389"
		Conf.Server.ServerPort = "28008"
		Conf.Server.ServerHost = "127.0.0.1"
		Conf.Server.ServerHandlerKey = "handler_key"

	}

	err = viper.Unmarshal(Conf)
	if err != nil {
		logs.Fatal(err.Error())
		os.Exit(1)
	}

}
