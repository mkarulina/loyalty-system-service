package config

import (
	"flag"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	serverAddress string `yaml:"SERVER_ADDRESS"`
	baseURL       string `yaml:"BASE_URL"`
	filePath      string `yaml:"FILE_STORAGE_PATH"`
	dbAddress     string `yaml:"DATABASE_DSN"`
}

func LoadConfig(path string) (config *Config, err error) {
	conf := &Config{}

	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
		return conf, err
	}

	flag.StringVar(&conf.serverAddress, "a", ":8080", "port to listen on")
	flag.StringVar(&conf.baseURL, "b", "http://localhost:8080/", "base short url")
	flag.StringVar(&conf.filePath, "f", "../urls.log", "file path for url saving")
	flag.StringVar(&conf.dbAddress, "d", "postgresql://localhost:5432/postgres", "data base address")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			viper.Set("SERVER_ADDRESS", &conf.serverAddress)
		case "b":
			viper.Set("BASE_URL", &conf.baseURL)
		case "f":
			viper.Set("FILE_STORAGE_PATH", &conf.filePath)
		case "d":
			viper.Set("DATABASE_DSN", &conf.dbAddress)
		}
	})

	return conf, err
}
