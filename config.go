package main

import (
	"log"
	"os"
)

type Config struct {
	FileServerPath string
}

var cfg Config

func (config *Config) init() {
	if p := os.Getenv("FILE_SERVER_PATH"); p != "" {
		config.FileServerPath = p
	} else {
		config.FileServerPath = "./frontend/dist/ff-frontend/browser"
	}

	log.Printf("[config] loaded: %+v\n", cfg)
}
