package configs

import (
	"fsync/client/global"

	"github.com/spf13/viper"
)

func LoadConfig(path string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv() // 允许环境变量覆盖

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&global.Configs); err != nil {
		return err
	}

	return nil
}
