package main

import (
	"fmt"
	"os"

	"github.com/designcomb/influenter-backend/internal/utils"
)

func main() {
	key, err := utils.GenerateEncryptionKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating encryption key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated ENCRYPTION_KEY (base64 encoded, 32 bytes):")
	fmt.Println(key)
	fmt.Println("\nYou can set it as an environment variable:")
	fmt.Printf("export ENCRYPTION_KEY=%s\n", key)
	fmt.Println("\nOr add it to your .env file:")
	fmt.Printf("ENCRYPTION_KEY=%s\n", key)
}

