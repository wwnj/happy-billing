package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	InfluxDB InfluxDBConfig `mapstructure:"influxdb"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"` // dev/test/prod
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`  // 秒
	WriteTimeout int    `mapstructure:"write_timeout"` // 秒
}

// DatabaseConfig PostgreSQL数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 分钟
	SSLMode         string `mapstructure:"ssl_mode"`          // disable/require/verify-ca/verify-full
}

// DSN 生成PostgreSQL连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Database, d.SSLMode,
	)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// Addr 生成Redis地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers      []string `mapstructure:"brokers"`
	GroupID      string   `mapstructure:"group_id"`
	AutoCommit   bool     `mapstructure:"auto_commit"`
	MaxRetries   int      `mapstructure:"max_retries"`
	RetryBackoff int      `mapstructure:"retry_backoff"` // 毫秒
	BatchSize    int      `mapstructure:"batch_size"`
	BatchTimeout int      `mapstructure:"batch_timeout"` // 毫秒
}

// InfluxDBConfig InfluxDB配置
type InfluxDBConfig struct {
	URL           string `mapstructure:"url"`
	Token         string `mapstructure:"token"`
	Org           string `mapstructure:"org"`
	Bucket        string `mapstructure:"bucket"`
	BatchSize     int    `mapstructure:"batch_size"`
	FlushInterval int    `mapstructure:"flush_interval"` // 秒
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`       // debug/info/warn/error
	Format     string `mapstructure:"format"`      // json/console
	Output     string `mapstructure:"output"`      // stdout/stderr/file
	FilePath   string `mapstructure:"file_path"`   // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"` // 最大备份数
	MaxAge     int    `mapstructure:"max_age"`     // 天
	Compress   bool   `mapstructure:"compress"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath("../../configs")
	}

	// 支持环境变量覆盖
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 设置全局配置
	globalConfig = &cfg

	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded, please call Load() first")
	}
	return globalConfig
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "dev" || c.App.Environment == "development"
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Environment == "prod" || c.App.Environment == "production"
}
