package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"kryptos/crypto" // Replace with your actual module path
	"kryptos/entry"  // Replace with your actual module path
)

const dataFileName = "kryptos_data.json"
const accountFileName = "kryptos_accounts.json"

type EncryptedData struct {
	Salt             []byte `json:"salt"`
	EncryptedEntries string `json:"entries"`
}

// GetDataFilePath returns the path to the data file in the user's config directory.
func GetDataFilePath(accountName string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "Kryptos") // App-specific directory
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		if err := os.MkdirAll(appDir, 0700); err != nil {
			return "", err // Return error if directory creation fails
		}
	}
	return filepath.Join(appDir, fmt.Sprintf("%s_%s", accountName, dataFileName)), nil
}

// LoadEncryptedData loads and decrypts data from the JSON file.
func LoadEncryptedData(accountName string, masterPassword string) ([]entry.PasswordEntry, error) {
	filePath, err := GetDataFilePath(accountName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []entry.PasswordEntry{}, nil // File doesn't exist, return empty slice
		}
		return nil, err
	}

	var encryptedDataStructure EncryptedData
	if err := json.Unmarshal(data, &encryptedDataStructure); err != nil {
		return nil, err
	}

	salt := encryptedDataStructure.Salt
	key, err := crypto.DeriveKeyFromPassword(masterPassword, salt)
	if err != nil {
		return nil, err
	}

	decryptedData, err := crypto.DecryptData(encryptedDataStructure.EncryptedEntries, key)
	if err != nil {
		return nil, err
	}

	var entries []entry.PasswordEntry
	if err := json.Unmarshal(decryptedData, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// SaveEncryptedData encrypts and saves data to the JSON file.
func SaveEncryptedData(accountName string, masterPassword string, entries []entry.PasswordEntry) error {
	filePath, err := GetDataFilePath(accountName)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	salt, err := crypto.GenerateSalt() // Generate a new salt on each save (more secure)
	if err != nil {
		return err
	}

	key, err := crypto.DeriveKeyFromPassword(masterPassword, salt)
	if err != nil {
		return err
	}

	encryptedData, err := crypto.EncryptData(jsonData, key)
	if err != nil {
		return err
	}

	encryptedDataStructure := EncryptedData{
		Salt:             salt,
		EncryptedEntries: encryptedData,
	}

	dataToSave, err := json.Marshal(encryptedDataStructure)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, dataToSave, 0600) // 0600 permissions - owner read/write only
}

// ExportEncryptedData exports the encrypted data to a file.
func ExportEncryptedData(accountName string, masterPassword string, exportPath string) error {
	filePath, err := GetDataFilePath(accountName)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return os.WriteFile(exportPath, data, 0600)
}

// ImportEncryptedData imports encrypted data from a file.
func ImportEncryptedData(accountName string, masterPassword string, importPath string) error {
	importData, err := os.ReadFile(importPath)
	if err != nil {
		return err
	}

	filePath, err := GetDataFilePath(accountName)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, importData, 0600)
}

// GetAccountFilePath returns the path to the data file in the user's config directory.
func GetAccountFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "Kryptos") // App-specific directory
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		if err := os.MkdirAll(appDir, 0700); err != nil {
			return "", err // Return error if directory creation fails
		}
	}
	return filepath.Join(appDir, accountFileName), nil
}

// LoadAccounts loads account names from a JSON file
func LoadAccounts() []string {
	filePath, err := GetAccountFilePath()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{} // Return empty slice
		}
		return nil
	}
	var accounts []string
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil
	}
	return accounts
}

// SaveAccounts saves account names to a JSON file.
func SaveAccounts(accounts []string) error {
	filePath, err := GetAccountFilePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(accounts)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600)
}
