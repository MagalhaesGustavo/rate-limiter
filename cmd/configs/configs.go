package configs

import (
	"github.com/spf13/viper"
)

type Conf struct {
	RedisHost          string `mapstructure:"DB_HOST"`
	RequestsByIp       int    `mapstructure:"REQUESTS_IP"`
	RequestsByToken    int    `mapstructure:"REQUESTS_TOKEN"`
	TimeBlockedByIp    int    `mapstructure:"TIME_BLOCKED_IP"`
	TimeBlockedByToken int    `mapstructure:"TIME_BLOCKED_TOKEN"`
	Port      		   string `mapstructure:"PORT"`
	AllowedToken       string `mapstructure:"ALLOWED_TOKEN"`
}

func LoadConfig(path string) (*Conf, error) {
	var configuration *Conf
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&configuration)
	if err != nil {
		panic(err)
	}

	return configuration, nil
}