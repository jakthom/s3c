package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	// Config File/Env Defaults
	S3C_CONFIG_PATH              string = "S3C_CONFIG_PATH"
	DEFAULT_HERCULES_CONFIG_PATH string = "s3c.yml"
	DEBUG                        string = "DEBUG"
	TRACE                        string = "TRACE"
	YAML_CONFIG_TYPE             string = "yaml"
)

type Origin struct {
	Type   string `json:"type"`   // One of "fs", "s3", "gcs", "r2"
	Bucket string `json:"bucket"` // The s3-compat bucket name
}

type Auth struct {
	KeyID  string `json:"keyId"`
	Secret string `json:"secret"`
}

type Config struct {
	Port   string `json:"port"`
	Origin `json:"origin"`
	Auth   `json:"auth"`
}

// Get configuration. If the specified file cannot be read fall back to sane defaults.
func GetConfig() (Config, error) {
	// Load app config from file
	confPath := os.Getenv(S3C_CONFIG_PATH)
	if confPath == "" {
		confPath = DEFAULT_HERCULES_CONFIG_PATH
	}
	log.Info().Msg("loading config from " + confPath)
	config := &Config{}
	// Try to get configuration from file
	viper.SetConfigFile(confPath)
	viper.SetConfigType(YAML_CONFIG_TYPE)
	err := viper.ReadInConfig()
	if err := viper.Unmarshal(config); err != nil {
		log.Error().Stack().Err(err)
	}
	return *config, err
}
