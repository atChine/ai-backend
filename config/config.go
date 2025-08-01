package config

import (
	"time"

	"github.com/spf13/viper"
)

// AppConfig 应用配置
var AppConfig Config

// Config 配置结构
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	AI     AIConfig     `mapstructure:"ai"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AIConfig AI服务配置
type AIConfig struct {
	ModelName    string        `mapstructure:"model_name"`
	Timeout      time.Duration `mapstructure:"timeout"`
	MaxTokens    int           `mapstructure:"max_tokens"`
	StreamBuffer int           `mapstructure:"stream_buffer"`
}

// Init 初始化配置
func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", 10)
	viper.SetDefault("server.write_timeout", 10)
	viper.SetDefault("ai.model_name", "default-model")
	viper.SetDefault("ai.timeout", 30)
	viper.SetDefault("ai.max_tokens", 1024)
	viper.SetDefault("ai.stream_buffer", 1024)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return viper.Unmarshal(&AppConfig)
}
