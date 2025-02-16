package config

import "os"

type Config struct {
	Address      string
	PostgresConn string
}

func GetConfig() *Config {
	config := &Config{
		Address:      os.Getenv("SERVER_ADDRESS"),
		PostgresConn: os.Getenv("POSTGRES_CONN"),
	}
	return config
}
