package config

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HostName string
	HideIP   bool
	Auth     []string
	Logging  []string
}

var defaults = map[string]interface{}{
	"hideIP":  false,
	"logging": []string{},
}

func init() {
	for key, item := range defaults {
		viper.SetDefault(key, item)
	}
}

func LoadConfig(path string) (*Config, error) {
	if path != "" {
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error loading configuration file: %w", err)
		}

	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &Config{
		HostName: viper.GetString("server.host"),
		HideIP:   viper.GetBool("hideIP"),
		Auth:     viper.GetStringSlice("server.auth"),
		Logging:  viper.GetStringSlice("logging"),
	}, nil
}

func (c *Config) CheckAuthString(auth string) bool {
	split := strings.Split(auth, " ")
	if len(split) != 2 || split[0] != "Basic" {
		return false
	}

	auth = split[1]
	decoded, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return false
	}
	var found bool
	for _, item := range c.Auth {
		if string(decoded) == item {
			found = true
			break
		}
	}
	return found
}

func (c *Config) HasAuth() bool {
	return len(c.Auth) > 0
}
