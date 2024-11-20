package main

import (
	"db-backup/models"
	"db-backup/utils"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Servers []models.ServerConfig `json:"servers"`
}

// Load environment variables from .env file
func loadEnv() error {
	return godotenv.Load(".env")
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
		key, err := utils.GenerateAESKey()
		if err != nil {
			logger.Fatalf("Error generating AES key: %v\n", err)
		}
		logger.Println("AES key generated successfully")

		// Loop through each server and back up its database
		var results []models.ResultMessage
		for _, server := range config.Servers {
			var aResult models.ResultMessage
			aResult.ServerName = server.Name
			logger.Printf("Starting backup for server: %s\n", server.Name)
			err := utils.BackupDatabase(server, logger, key, outputFolder)
			if err != nil {
				logger.Printf("Error backing up server '%s': %v\n", server.Name, err)
				aResult.Success = false
				aResult.Message = fmt.Sprintf("Error backing up server '%s': %v", server.Name, err)
			} else {
				aResult.Success = true
				aResult.Message = fmt.Sprintf("Backing up server '%s': Success", server.Name)
			}

			results = append(results, aResult)
		}
		errDiscord := utils.PostMessage(results)
		if errDiscord != nil {
			logger.Println("Post to Discord : Fail - ", errDiscord)
		} else {
			logger.Println("Post to Discord : Success")
		}
		logger.Println("Backup process completed")
	} else if *decryptFlag {
		if *encryptedFile == "" {
			fmt.Println("For decryption, param -file must be specified")
			return
		}

		key, err := utils.GenerateAESKey()
		if err != nil {
			fmt.Printf("Error generating AES key: %v\n", err)
			return
		}

		// output file
		outputFile := strings.Replace(*encryptedFile, ".enc", "", 1)

		err = utils.DecryptFile(*encryptedFile, outputFile, key)
		if err != nil {
			fmt.Printf("Error decrypting file: %v\n", err)
		} else {
			fmt.Println("Decryption successful")
		}
	} else {
		fmt.Println("Invalid command. Use -backup or -decrypt.")
	}
}
