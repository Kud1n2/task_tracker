package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env           string `yaml:"env" env-default:"local"`
	StorageConfig `yaml:"storage"`
	CacheConfig   `yaml:"cache"`
	HTTPServer    `yaml:"http_server"`
	JWT           `yaml:"jwt"`
}

type HTTPServer struct {
	Address         string        `yaml:"address" env-default:"localhost:8080"`
	Timeout         time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	DatabaseTimeout time.Duration `yaml:"database_timeout" env:"DB_TIMEOUT" env-default:"10s"`
}

type StorageConfig struct {
	StoragePath     string        `yaml:"storage_path" env-required:"true"`
	MaxOpenConns    int           `yaml:"max_open_connections" env-default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_connections" env-default:"12"`
	ConnMaxLifetime time.Duration `yaml:"max_connection_lifetime" env-default:"300s"`
	ConnMaxIdleTime time.Duration `yaml:"max_idle_connection_lifetime" env-default:"60s"`
}

type CacheConfig struct {
	RedisAddr    string `yaml:"redis_addr" env:"REDIS_ADDR" env-default:"localhost:6379"`
	PoolSize     int    `yaml:"redis_pool_size" env-default:"10"`
	MinIdleConns int    `yaml:"min_idle_connections" env-default:"3"`
}

type JWT struct {
	Secret string
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("Empty configPath")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s doesn't exist", configPath)
	}

	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatal("Error reading config file")
	}

	return &config
}
