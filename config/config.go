package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HostName string
	HideIP   bool
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
		HideIP:   viper.GetBool("hide_ip"),
	}, nil
}
