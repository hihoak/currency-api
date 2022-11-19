package config

import (
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"log"
	"os"
	"time"
)

type LoggerSection struct {
	LogLevel string `default:"info" env:"LOG_LEVEL"`
}

type ServerSection struct {
	Address string `default:"127.0.0.1:8000" env:"ADDRESS"`
}

type DatabaseSection struct {
	Host     string `default:"localhost" env:"HOST"`
	Port     string `default:"6432" env:"PORT"`
	User     string `default:"postgres" env:"USER"`
	Password string `default:"" env:"PASSWORD"`
	DBName   string `default:"postgres" env:"DB_NAME"`
	ConnectionTimeout time.Duration `default:"2s" env:"CONNECTION_TIMEOUT"`
	OperationTimeout  time.Duration `default:"2s" env:"OPERATION_TIMEOUT"`
}

type Config struct {
	Logger        LoggerSection
	Server        ServerSection
	Database	  DatabaseSection
}

func New(configPath string) *Config {
	var cfg Config

	if _, err := os.Stat(configPath); err != nil {
		log.Println("fail to get config file, continue without it:", err)
	}

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipFlags: true,

		EnvPrefix: "CURRENCY_API",
		Files:     []string{configPath},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})

	if loadErr := loader.Load(); loadErr != nil {
		log.Fatal("failed to load configuration", loadErr)
	}

	return &cfg
}
