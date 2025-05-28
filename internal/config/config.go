package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server ServerConfig
	Redis  RedisConfig
}

type ServerConfig struct {
	Port string
}

type RedisConfig struct {
	Addrs    []string
	Password string
	DB       int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: os.Getenv("PORT"),
		},
		Redis: RedisConfig{
			Addrs:    parseList(os.Getenv("REDIS_ADDRS")),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       coerceInt(os.Getenv("REDIS_DB")),
		},
	}
}

func coerceInt(s string) int {
	if s == "" {
		return 0
	}
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return value
}

func parseList(s string) []string {
	lst := strings.Split(s, ",")
	for i := range lst {
		lst[i] = strings.TrimSpace(lst[i])
	}
	return lst
}
