package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	// TemplateExt stores the extension used for the template files
	TemplateExt = ".gohtml"

	// StaticDir stores the name of the directory that will serve static files
	StaticDir = "static"

	// StaticPrefix stores the URL prefix used when serving static files
	StaticPrefix = "files"
)

type environment string

const (
	// EnvLocal represents the local environment
	EnvLocal environment = "local"

	// EnvTest represents the test environment
	EnvTest environment = "test"

	// EnvDevelop represents the development environment
	EnvDevelop environment = "dev"

	// EnvStaging represents the staging environment
	EnvStaging environment = "staging"

	// EnvQA represents the qa environment
	EnvQA environment = "qa"

	// EnvProduction represents the production environment
	EnvProduction environment = "prod"
)

// SwitchEnvironment sets the environment variable used to dictate which environment the application is
// currently running in.
// This must be called prior to loading the configuration in order for it to take effect.
func SwitchEnvironment(env environment) {
	if err := os.Setenv("REPLC_APP_ENVIRONMENT", string(env)); err != nil {
		panic(err)
	}
}

type (
	// Config stores complete configuration
	Config struct {
		Web                WebConfig
		App                AppConfig
		OrchestratorConfig OrchestratorConfig
		Database           DatabaseConfig
		S3                 S3Config
		K8S                K8SConfig
	}

	// WebConfig stores HTTP configuration
	WebConfig struct {
		Hostname string
		Port     uint16
	}

	// AppConfig stores application configuration
	AppConfig struct {
		Name        string
		Environment environment
	}

	// OrchestratorConfig stores the cache configuration
	OrchestratorConfig struct {
		Hostname string
		Port     uint16
	}

	// DatabaseConfig stores the database configuration
	DatabaseConfig struct {
		Hostname string
		Port     uint16
		User     string
		Password string
		Database string
	}

	S3Config struct {
		Accesskey string
		Secretkey string
		Region    string
	}

	K8SConfig struct {
		Kubeconfigpath  string
		Serviceyamlpath string
		Namespace       string
	}
)

// GetConfig loads and returns configuration
func GetConfig() (Config, error) {
	var c Config

	// Load the config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	// Load env variables
	viper.SetEnvPrefix("replc")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return c, err
	}

	if err := viper.Unmarshal(&c); err != nil {
		return c, err
	}

	return c, nil
}
