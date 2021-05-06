package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	HostName       string
	HideIP         bool
	Auth           []string
	Logging        []string
	Timeout        time.Duration
	AllowedIP      []string
	LivenessStatus int
	LivenessBody   string
	LivenessPath   string
}

var defaults = map[string]interface{}{
	"hideIP":             false,
	"server.timeout":     time.Second * 30,
	"server.health.path": "/",
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
		HostName:       viper.GetString("server.host"),
		HideIP:         viper.GetBool("hideIP"),
		Auth:           viper.GetStringSlice("server.auth"),
		Logging:        viper.GetStringSlice("logging"),
		Timeout:        viper.GetDuration("server.timeout"),
		AllowedIP:      viper.GetStringSlice("allowedIP"),
		LivenessStatus: viper.GetInt("server.health.status"),
		LivenessBody:   viper.GetString("server.health.body"),
		LivenessPath:   viper.GetString("server.health.path"),
	}, nil
}

func (c *Config) Validate() error {
	if c.HostName == "" {
		return errors.New("server.host value can not be empty")
	}
	return nil
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

func (c *Config) ShouldAllowIP(addr string) bool {
	if len(c.AllowedIP) < 1 {
		return true
	}
	cutPos := len(addr)
	if pos := strings.Index(addr, ":"); pos != -1 {
		cutPos = pos
	}
	addr = addr[:cutPos]

	found := false
	for _, item := range c.AllowedIP {
		if item == addr {
			found = true
			break
		}
	}
	return found
}
