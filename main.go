package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
}

type Config struct {
	Servers []ServerConfig `json:"servers"`
}

// Load environment variables from .env file
func loadEnv() error {
	return godotenv.Load(".env")
}

// Generate random AES key
func generateAESKey() ([]byte, error) {
	key := make([]byte, 32) // 256-bit key
	keyString := os.Getenv("KEY")
	copy(key, keyString)

	return key, nil
}

// Encrypt file content
func encryptFile(inputFile string, outputFile string, key []byte) error {
	plainText, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading input file: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error creating cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("error creating GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("error generating nonce: %v", err)
	}

	cipherText := gcm.Seal(nonce, nonce, plainText, nil)
	err = ioutil.WriteFile(outputFile, cipherText, 0644)
	if err != nil {
		return fmt.Errorf("error writing encrypted file: %v", err)
	}
	return nil
}

// Decrypt file content
func decryptFile(inputFile string, outputFile string, key []byte) error {
	cipherText, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading encrypted file: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error creating cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("error creating GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return fmt.Errorf("error decrypting file: %v", err)
	}

	err = ioutil.WriteFile(outputFile, plainText, 0644)
	if err != nil {
		return fmt.Errorf("error writing decrypted file: %v", err)
	}
	return nil
}

func backupDatabase(server ServerConfig, logger *log.Logger, key []byte, outputFolder string) error {
	// Create a timestamped backup file name
	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("db_%s_backup_%s.sql", server.Name, timestamp)
	outputFolder = outputFolder + "/" + server.Name
	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		fmt.Printf("Error creating output folder: %v\n", err)
		return err
	}

	encryptedFileName := filepath.Join(outputFolder, fmt.Sprintf("db_%s_backup_%s.sql.enc", server.Name, timestamp))

	// Construct mysqldump command
	cmd := exec.Command("mysqldump",
		"-u"+server.User,
		"-p"+server.Password,
		"-h"+server.Host,
		"-P"+fmt.Sprintf("%d", server.Port),
		server.Database,
	)

	// Create the backup file
	backupFile, err := os.Create(backupFileName)
	if err != nil {
		return fmt.Errorf("error creating backup file: %v", err)
	}
	defer backupFile.Close()

	// Redirect command output to the file
	cmd.Stdout = backupFile

	// Run the command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing mysqldump: %v", err)
	}

	logger.Printf("Backup successful for server '%s'. File: %s\n", server.Name, backupFileName)

	// Encrypt the backup file
	err = encryptFile(backupFileName, encryptedFileName, key)
	if err != nil {
		return fmt.Errorf("error encrypting backup file: %v", err)
	}

	// Remove the unencrypted file
	err = os.Remove(backupFileName)
	if err != nil {
		logger.Printf("Warning: Unable to delete unencrypted backup file: %v\n", err)
	}

	logger.Printf("Encrypted backup saved as: %s\n", encryptedFileName)
	return nil
}

func loadConfig(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

func main() {
	backupFlag := flag.Bool("backup", false, "Start the backup process")
	decryptFlag := flag.Bool("decrypt", false, "Decrypt a backup file")
	encryptedFile := flag.String("file", "", "Path to the encrypted file for decryption")

	flag.Parse()

	// Load environment variables
	err := loadEnv()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	// Get log and output folder from environment variables
	logFolder := os.Getenv("LOG_FOLDER")
	outputFolder := os.Getenv("OUTPUT_FOLDER")
	configFile := os.Getenv("CONFIG_FILE")

	// Ensure log folder exists
	err = os.MkdirAll(logFolder, 0755)
	if err != nil {
		fmt.Printf("Error creating log folder: %v\n", err)
		return
	}

	// Ensure output folder exists
	err = os.MkdirAll(outputFolder, 0755)
	if err != nil {
		fmt.Printf("Error creating output folder: %v\n", err)
		return
	}

	// Load configuration
	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	if *backupFlag {
		// Create log file
		logFileName := filepath.Join(logFolder, time.Now().Format("backup_log_20060102_150405.log"))
		logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error creating log file: %v\n", err)
			return
		}
		defer logFile.Close()

		logger := log.New(logFile, "", log.LstdFlags)
		logger.Println("Backup process started")

		// Generate an AES key (in production, save it securely)
		key, err := generateAESKey()
		if err != nil {
			logger.Fatalf("Error generating AES key: %v\n", err)
		}
		logger.Println("AES key generated successfully")

		// Loop through each server and back up its database
		for _, server := range config.Servers {
			logger.Printf("Starting backup for server: %s\n", server.Name)
			err := backupDatabase(server, logger, key, outputFolder)
			if err != nil {
				logger.Printf("Error backing up server '%s': %v\n", server.Name, err)
			}
		}

		logger.Println("Backup process completed")
	} else if *decryptFlag {
		if *encryptedFile == "" {
			fmt.Println("For decryption, both -file must be specified")
			return
		}

		key, err := generateAESKey()
		if err != nil {
			fmt.Printf("Error generating AES key: %v\n", err)
			return
		}

		// output file
		outputFile := strings.Replace(*encryptedFile, ".enc", "", 1)

		err = decryptFile(*encryptedFile, outputFile, key)
		if err != nil {
			fmt.Printf("Error decrypting file: %v\n", err)
		} else {
			fmt.Println("Decryption successful")
		}
	} else {
		fmt.Println("Invalid command. Use -backup or -decrypt.")
	}
}
