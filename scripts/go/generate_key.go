package main

import (
	"crypto/rand"
	"fmt"
	"os"
)

func generateAES256GCMKey() ([]byte, error) {
	key := make([]byte, 32) // 256 bits = 32 bytes
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func saveKeyToFile(key []byte, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(key)
	return err
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: generate_aes256_gcm_key <output_file>")
		os.Exit(1)
	}

	fileName := os.Args[1]
	key, err := generateAES256GCMKey()
	if err != nil {
		fmt.Printf("Error generating AES-256-GCM key: %v\n", err)
		os.Exit(1)
	}

	err = saveKeyToFile(key, fileName)
	if err != nil {
		fmt.Printf("Error saving key to file: %v\n", err)
		os.Exit(1)
	}
}
