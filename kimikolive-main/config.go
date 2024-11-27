package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Model         string `json:"model"`
	Voice         string `json:"voice"`
	StreamKey     string `json:"stream_key"`
	SolanaNetwork string `json:"solana_network"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
} 