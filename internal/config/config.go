package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

const (
	ConfigFile                string = "credentials"
	ConfigFolder              string = "follower"
	EnvSessionBearerToken     string = "FOLLOWER_SESSION_TOKEN"
	EnvSessionTokenExpiryTime string = "FOLLOWER_SESSION_EXPIRY"
	EnvSessionTenant          string = "FOLLOWER_TENANT"
	EnvSessionApiUrl          string = "FOLLOWER_SESSION_API_URL"
	AtlassianApiVersion       string = "3"
)

func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, ConfigFolder), nil
}

func FilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, ConfigFile), nil
}

func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}

	return dir, nil
}

func LoadConfiguration() (*Configuration, error) {
	path, err := FilePath()
	if err != nil {
		return nil, err
	}

	var conf Configuration

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &conf, nil
		}

		return nil, err
	}

	if err := ini.MapTo(&conf, path); err != nil {
		return nil, err
	}

	return &conf, err
}

func SaveConfiguration(conf *Configuration) error {
	var err error

	if conf == nil {
		return errors.New("configuration is nil")
	}

	if _, err := EnsureDir(); err != nil {
		return err
	}

	cfg := ini.Empty()
	if err := ini.ReflectFrom(cfg, conf); err != nil {
		return err
	}

	path, err := FilePath()
	if err != nil {
		return err
	}

	return cfg.SaveTo(path)
}
