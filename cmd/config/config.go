package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ConfigName = "config.yaml"
)

// Default folder structure of compiled program will be look like this
// data/
// 		santex.db
// 		uploads/

type Config struct {
	DataDir string `yaml:"DataDirectory,omitempty"`
	ServerConfig
	DBConfig
	LogConfig
}

type LogConfig struct {
	LogFile  string `yaml:"LogFile,omitempty"`
	LogLevel string `yaml:"LogLevel,omitempty"`
}

type ServerConfig struct {
	// ServerPort is a port where server is running. Default is 0 and port is choosing by OS.
	ServerPort int    `yaml:"ServerPort,omitempty"`
	UploadsDir string `yaml:"UploadsDirectory,omitempty"`
}

type DBConfig struct {
	DBName string `yaml:"DatabaseName,omitempty"`
	DBPath string `yaml:"DatabasePath,omitempty"`
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
	if cfg.DataDir == "" {
		cfg.DataDir = "data"
		changed = true
	}

	if cfg.DBName == "" {
		cfg.DBName = "database.db"
		changed = true
	}

	if cfg.DBPath == "" {
		cfg.DBPath = filepath.Join(cfg.DataDir, cfg.DBName)
		changed = true
	}

	if cfg.UploadsDir == "" {
		cfg.UploadsDir = filepath.Join(cfg.DataDir, "uploads")
		changed = true
	}

	if cfg.LogFile == "" {
		cfg.LogFile = filepath.Join(cfg.DataDir, fmt.Sprintf("%s_log.log",
			strings.ReplaceAll(time.Now().Format(time.DateTime), ":", "-")))
		changed = true
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "INFO"
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
