package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	return Write(*c)
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	jsonFile, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config

	err = json.Unmarshal(jsonFile, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getConfigFilePath() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(userHomeDir, configFileName)
	return fullPath, nil
}

func Write(cfg Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	jsonFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}
