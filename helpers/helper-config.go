package helpers

import (
	"os"
	"path"
	"github.com/spf13/viper"
	"fmt"
)

// GetConfigPath retrieves the user-specific config directory and constructs the config file path.
func GetConfigPath() (string, error) {
	xampressPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to get user config directory: %w", err)
	}
	return path.Join(xampressPath, ".xampress", "config.yaml"), nil
}



func LoadConfig() error {
	configPath, _ := GetConfigPath()
	viper.AddConfigPath(path.Dir(configPath))
	viper.SetConfigName(path.Base(configPath[:len(configPath)-len(path.Ext(configPath))])) 
	viper.SetConfigType(path.Ext(configPath)[1:]) 

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("unable to read config file from %s: %w", configPath, err)
	}

	return nil
}


func Get(key string) (interface{}, error) {
	if err := LoadConfig(); err != nil {
		return nil, err
	}

	if !viper.IsSet(key) {
		return nil, fmt.Errorf("key %s not found in config", key)
	}

	return viper.Get(key), nil
}


func Set(key string, value interface{}) error {
	
	if err := LoadConfig(); err != nil {
		return err
	}

	viper.Set(key, value)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("unable to write config file: %w", err)
	}

	return nil
}

// GetAll retrieves all key-value pairs from the configuration file.
func GetAll() (map[string]any, error) {
	if err := LoadConfig(); err != nil {
		return nil, err
	}
	return viper.AllSettings(), nil
}
