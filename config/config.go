package config

import (
	"flag"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	runAddress     string `yaml:"RUN_ADDRESS"`
	dbAddress      string `yaml:"DATABASE_URI"`
	accrualAddress string `yaml:"ACCRUAL_SYSTEM_ADDRESS"`
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

	flag.StringVar(&conf.runAddress, "a", ":8080", "port to listen on")
	flag.StringVar(&conf.dbAddress, "d", "postgresql://localhost:5432/postgres", "data base address")
	flag.StringVar(&conf.accrualAddress, "r", ":8090", "accrual system address")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			viper.Set("RUN_ADDRESS", &conf.runAddress)
		case "d":
			viper.Set("DATABASE_URI", &conf.dbAddress)
		case "r":
			viper.Set("ACCRUAL_SYSTEM_ADDRESS", &conf.accrualAddress)
		}
	})

	return conf, err
}
