package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"db-backup/models"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Generate random AES key
func GenerateAESKey() ([]byte, error) {
	key := make([]byte, 32) // 256-bit key
	keyString := os.Getenv("KEY")
	copy(key, keyString)

	return key, nil
}

// Encrypt file content
func EncryptFile(inputFile string, outputFile string, key []byte) error {
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
func DecryptFile(inputFile string, outputFile string, key []byte) error {
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

func BackupDatabase(server models.ServerConfig, logger *log.Logger, key []byte, outputFolder string) error {
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
	args := []string{
		"-u" + server.User,
		"-p" + server.Password,
		"-h" + server.Host,
		"-P" + fmt.Sprintf("%d", server.Port),
	}

	if server.Tables != nil {
		// dump specify table
		args = append(args, "--tables")       // --tables
		args = append(args, server.Database)  // database
		args = append(args, server.Tables...) // ...tables
		for _, table := range server.Tables {
			logger.Printf("--tables: %s\n", table)
		}
	} else if server.IgnoredTables != nil {
		// ignore specify table
		args = append(args, server.Database) // database
		for _, table := range server.IgnoredTables {
			args = append(args, "--ignore-table="+table) // --ignore-table
			logger.Printf("--ignore-table: %s\n", table)
		}
	} else {
		args = append(args, server.Database) // database
	}
	cmd := exec.Command("mysqldump", args...)

	// Create the backup file
	backupFile, err := os.Create(backupFileName)
	if err != nil {
		return fmt.Errorf("error creating backup file: %v", err)
	}
	// defer backupFile.Close()
	// Properly close the file
	defer func() {
		if err := backupFile.Close(); err != nil {
			logger.Printf(`error closing file: %v`, err)
		}
	}()

	// Redirect command output to the file
	cmd.Stdout = backupFile

	// Run the command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing mysqldump: %v", err)
	}

	logger.Printf("Backup successful for server '%s'. File: %s\n", server.Name, backupFileName)

	// Encrypt the backup file
	err = EncryptFile(backupFileName, encryptedFileName, key)
	if err != nil {
		return fmt.Errorf("error encrypting backup file: %v", err)
	}

	// Remove the unencrypted file
	err = os.Remove(backupFileName)
	if err != nil {
		logger.Printf("Warning: Unable to delete unencrypted backup file: %v\n", err)
	} else {
		//delete file success
		logger.Printf("Deleting unencrypted backup file successful\n")
	}

	logger.Printf("Encrypted backup saved as: %s\n", encryptedFileName)
	return nil
}
