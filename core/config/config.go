package config

import (
	"errors"
	"strings"
	"sync"

	"github.com/apex/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Name string
}

func (c *Config) Init() {
	// Set default values
	c.setDefaults()

	// config
	if c.Name != "" {
		viper.SetConfigFile(c.Name) // if config file is set, load it accordingly
	} else {
		viper.AddConfigPath("./conf") // if no config file is set, load by default
		viper.SetConfigName("config")
	}

	// config type as yaml
	viper.SetConfigType("yaml")

	// auto env
	viper.AutomaticEnv()

	// env prefix
	viper.SetEnvPrefix("CRAWLAB")

	// replacer
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	// read in config
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Warn("No config file found. Using default values.")
		}
	}

	// init log level
	c.initLogLevel()
}

func (c *Config) WatchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed: %s", e.Name)
	})
}

func (c *Config) setDefaults() {
	viper.SetDefault("mongo.host", "localhost")
	viper.SetDefault("mongo.port", 27017)
	viper.SetDefault("mongo.db", "crawlab_test")
	viper.SetDefault("mongo.username", "")
	viper.SetDefault("mongo.password", "")
	viper.SetDefault("mongo.authSource", "admin")
}

func (c *Config) initLogLevel() {
	// set log level
	logLevel := viper.GetString("log.level")
	l, err := log.ParseLevel(logLevel)
	if err != nil {
		l = log.InfoLevel
	}
	log.SetLevel(l)
}

func newConfig() *Config {
	return &Config{}
}

var _config *Config
var _configOnce sync.Once

func GetConfig() *Config {
	_configOnce.Do(func() {
		_config = newConfig()
		_config.Init()
	})
	return _config
}

func InitConfig() {
	// config instance
	c := GetConfig()

	// watch config change and load responsively
	c.WatchConfig()
}
