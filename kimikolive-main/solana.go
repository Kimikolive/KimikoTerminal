package main

import (
	"errors"
	"os"
)

type SolanaClient struct {
	PrivateKey string
	Network    string
}

func NewSolanaClient(network string) (*SolanaClient, error) {
	privateKey := os.Getenv("SOLANA_PRIVATE_KEY")
	if privateKey == "" {
		return nil, errors.New("SOLANA_PRIVATE_KEY not set")
	}

	return &SolanaClient{
		PrivateKey: privateKey,
		Network:    network,
	}, nil
} 