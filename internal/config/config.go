package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string         `mapstructure:"environment"`
	Version     string         `mapstructure:"version"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Redis       RedisConfig    `mapstructure:"redis"`
	Temporal    TemporalConfig `mapstructure:"temporal"`
	Kubernetes  K8sConfig      `mapstructure:"kubernetes"`
	Vault       VaultConfig    `mapstructure:"vault"`
	JWT         JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

type TemporalConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	TaskQueue string `mapstructure:"task_queue"`
}

type K8sConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig"`
	InCluster  bool   `mapstructure:"in_cluster"`
	Namespace  string `mapstructure:"namespace"`
}

type VaultConfig struct {
	Address  string `mapstructure:"address"`
	Token    string `mapstructure:"token"`
	Path     string `mapstructure:"path"`
	AuthType string `mapstructure:"auth_type"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/runbook-engine")

	// Set defaults
	viper.SetDefault("environment", "development")
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 15)
	viper.SetDefault("server.write_timeout", 15)
	viper.SetDefault("server.idle_timeout", 60)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("temporal.host", "localhost")
	viper.SetDefault("temporal.port", 7233)
	viper.SetDefault("temporal.namespace", "default")
	viper.SetDefault("temporal.task_queue", "runbook-queue")
	viper.SetDefault("kubernetes.in_cluster", false)
	viper.SetDefault("kubernetes.namespace", "default")
	viper.SetDefault("vault.address", "http://localhost:8200")
	viper.SetDefault("vault.path", "secret")
	viper.SetDefault("vault.auth_type", "token")
	viper.SetDefault("jwt.expiration", 3600)

	// Environment variable overrides
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults and environment variables
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables for sensitive data
	if dbPassword := os.Getenv("DATABASE_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}
	if vaultToken := os.Getenv("VAULT_TOKEN"); vaultToken != "" {
		config.Vault.Token = vaultToken
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
	}

	return &config, nil
}
