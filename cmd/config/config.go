package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	ConfigName = "config.yaml"
)

// Default folder structure of compiled program will be look like this
// binary
// config.yaml
// base_directory/
// 		database_name.db
// 		uploads/

type Config struct {
	BaseDirectory string       `yaml:"base_directory,omitempty"`
	ServerCfg     ServerConfig `yaml:"server,omitempty"`
	DBCfg         DBConfig     `yaml:"database,omitempty"`
	LogCfg        LogConfig    `yaml:"log,omitempty"`
}

type LogConfig struct {
	LogLevel string `yaml:"log_level,omitempty"`
}

type ServerConfig struct {
	// ServerPort is a port where server is running. Default is 0 and port is choosing by OS.
	ServerPort int    `yaml:"server_port"`
	UploadsDir string `yaml:"uploads_directory,omitempty"`
}

type DBConfig struct {
	DBName string `yaml:"database_name,omitempty"`
}

func NewConfig(cfgPath string) (Config, error) {
	file, err := os.Open(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {

			file, err = setupConfigFile(cfgPath)
			if err != nil {
				return Config{}, err
			}

		} else {
			return Config{}, err
		}
	}
	cfg, err := parseConfig(file)
	_ = file.Close()

	if cfg.setDefaults() {
		err = cfg.Save()
	}
	return cfg, err
}

func parseConfig(file *os.File) (Config, error) {
	cfg := Config{}

	buf, err := io.ReadAll(file)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, err
}

func setupConfigFile(cfgPath string) (*os.File, error) {
	cfg := Config{}
	cfg.setDefaults()

	file, err := os.OpenFile(cfgPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("can't open new config file: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("can't marshal default config: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return nil, fmt.Errorf("can't save newly created config file: %w", err)
	}
	return file, err
}

func (cfg *Config) setDefaults() bool {
	var changed bool
	if cfg.BaseDirectory == "" {
		cfg.BaseDirectory = "data"
		changed = true
	}

	if cfg.DBCfg.DBName == "" {
		cfg.DBCfg.DBName = "database.db"
		changed = true
	}

	if cfg.ServerCfg.UploadsDir == "" {
		cfg.ServerCfg.UploadsDir = "uploads"
		changed = true
	}

	if cfg.LogCfg.LogLevel == "" {
		cfg.LogCfg.LogLevel = "INFO"
		changed = true
	}

	return changed
}

func (cfg *Config) Save() error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("can't save config: %w", err)
	}
	err = os.WriteFile(ConfigName, data, 0644)
	if err != nil {
		return fmt.Errorf("can't save config: %w", err)
	}
	return nil
}

func (cfg *Config) UpdateConfig(newCfg Config) error {
	cfg = &newCfg
	return cfg.Save()
}

func (cfg *Config) ResetConfig() error {
	defaultCfg := &Config{}
	defaultCfg.setDefaults()
	cfg = defaultCfg
	return cfg.Save()
}
