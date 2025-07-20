package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	var user_config Config
	file_path, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("file not found: %v", err)
	}
	data, readError := os.ReadFile(file_path)
	if readError != nil {
		return Config{}, fmt.Errorf("error reading file: %v", readError)
	}
	unmarshallError := json.Unmarshal(data, &user_config)
	if unmarshallError != nil {
		return Config{}, fmt.Errorf("error unmarshalling data: %v", unmarshallError)
	}
	return user_config, nil

}

func SetUser(cfg Config, user_name string) error {
	cfg.CurrentUserName = user_name
	err := write(cfg)
	if err != nil {
		return fmt.Errorf("error setting user: %v", err)
	}
	return nil
}

func getConfigFilePath() (string, error) {
	home_path, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("errer getting home path: %v", err)
	}
	file_path := filepath.Join(home_path, configFileName)
	return file_path, nil
}

func write(cfg Config) error {
	file_path, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("error getting file path: %v", err)
	}
	cfg_marshalled, marshalErr := json.Marshal(cfg)
	if marshalErr != nil {
		return fmt.Errorf("error marshalling struct: %v", marshalErr)
	}
	err = os.WriteFile(file_path, cfg_marshalled, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	return nil
}
