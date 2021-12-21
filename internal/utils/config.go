package util

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Port           	int           `mapstructure:"PORT"`
	DbHost        	string        `mapstructure:"DB_HOST"`
	DbPort         	int           `mapstructure:"DB_PORT"`
	DbUser       	string        `mapstructure:"DB_USER"`
	DbPassword     	string        `mapstructure:"DB_PASSWORD"`
	DbDatabase     	string        `mapstructure:"DB_DATABASE"`
	DbSchema     	string        `mapstructure:"DB_SCHEMA"`
	LotusToken  	string        `mapstructure:"LOTUS_TOKEN"`
	LotusAddress  	string        `mapstructure:"LOTUS_ADDR"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	fmt.Println("utils config load")
	err = viper.Unmarshal(&config)
	return
}