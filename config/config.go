package config

import (
	"errors"
	"fmt"
	validator "github.com/asaskevich/govalidator"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	defaultLogLevel = "DEBUG"
	defaultLookback = 24
)

type Config struct {
	Log struct {
		Level string `yaml:"level" env:"LOG_LEVEL"`
	} `yaml:"log"`

	Tables []struct {
		Name      string `yaml:"name" env:"TABLE_NAME" valid:"minstringlength(3)"`
		Retention string `yaml:"retention" env:"TABLE_RETENTION" valid:"minstringlength(2)"`
	} `yaml:"tables"`

	Microsoft struct {
		AppID          string `yaml:"app_id" env:"MS_APP_ID" valid:"minstringlength(3)"`
		SecretKey      string `yaml:"secret_key" env:"MS_SECRET_KEY" valid:"minstringlength(3)"`
		TenantID       string `yaml:"tenant_id" env:"MS_TENANT_ID" valid:"minstringlength(3)"`
		SubscriptionID string `yaml:"subscription_id" env:"MS_SUB_ID" valid:"minstringlength(3)"`
		ResourceGroup  string `yaml:"resource_group" env:"MS_RSG_ID" valid:"minstringlength(3)"`
		WorkspaceName  string `yaml:"workspace_name" env:"MS_WS_NAME" valid:"minstringlength(3)"`
	} `yaml:"microsoft"`
}

func (c *Config) Validate() error {
	if c.Log.Level == "" {
		c.Log.Level = defaultLogLevel
	}

	if len(c.Tables) == 0 {
		return errors.New("no tables specified")
	}

	if valid, err := validator.ValidateStruct(c); !valid || err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	return nil
}

func (c *Config) Load(path string) error {
	if path != "" {
		configBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to load configuration file at '%s': %v", path, err)
		}

		if err = yaml.Unmarshal(configBytes, c); err != nil {
			return fmt.Errorf("failed to parse configuration: %v", err)
		}
	}

	if err := envconfig.Process("", c); err != nil {
		return fmt.Errorf("could not load environment: %v", err)
	}

	return nil
}
