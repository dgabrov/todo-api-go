package start

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/dgb9/todo-api/internal/data"
	"log/slog"
	"os"
	"strings"
)

func loadConfig() (data.Config, error) {
	res := data.Config{}

	// get first file
	configFile := os.Getenv("CONFIG_FILE")
	configFile = strings.TrimSpace(configFile)

	if len(configFile) == 0 {
		return res, errors.New("CONFIG_FILE environment variable not present")
	}

	passwordFile := os.Getenv("PASSWORD_FILE")
	passwordFile = strings.TrimSpace(passwordFile)

	if len(passwordFile) == 0 {
		return res, errors.New("PASSWORD_FILE environment variable not present")
	}
	// check the files exist
	_, err := os.Stat(configFile)
	if err != nil {
		return res, err
	}

	_, err = os.Stat(passwordFile)
	if err != nil {
		return res, err
	}

	// get password file
	var config data.Config
	var secret data.SecretConfig

	err = loadConfigFile(&config, configFile)
	if err != nil {
		return res, err
	}

	err = loadConfigFile(&secret, passwordFile)
	if err != nil {
		return res, err
	}

	// all is good, will populate the value for the password in the main config data
	dbConfig := config.Db
	slog.Info(fmt.Sprintf("secret loaded password is: %s", secret.Password))
	dbConfig.Password = secret.Password

	// let's log the config information
	slog.Info("System configuration")
	slog.Info(fmt.Sprintf("database user: %s", dbConfig.User))
	slog.Info(fmt.Sprintf("database machine: %s", dbConfig.Machine))
	slog.Info(fmt.Sprintf("database db: %s", dbConfig.Database))
	slog.Info(fmt.Sprintf("database port: %d", dbConfig.Port))

	passwordMessage := "empty"
	if len(dbConfig.Password) > 0 {
		passwordMessage = "filled out"
	}

	slog.Info(fmt.Sprintf("database password status: %s", passwordMessage))

	// regular server information
	serverConfig := config.Server
	slog.Info(fmt.Sprintf("server address: %s", serverConfig.Address))
	slog.Info(fmt.Sprintf("server authServerUrl: %s", serverConfig.AuthServerUrl))
	slog.Info(fmt.Sprintf("server right: %s", serverConfig.Right))
	slog.Info(fmt.Sprintf("server storageFolder: %s", serverConfig.StorageFolder))

	return config, nil
}

func loadConfigFile(pData any, fileName string) error {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, pData)
}
