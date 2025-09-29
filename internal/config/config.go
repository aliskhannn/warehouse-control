package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/wb-go/wbf/zlog"
)

type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	JWT      JWT      `mapstructure:"jwt"`
}

// Server holds HTTP server-related configuration.
type Server struct {
	HTTPPort     string        `mapstructure:"http_port"` // HTTP port to listen on
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// Database holds database master and slave configuration.
type Database struct {
	Master DatabaseNode   `mapstructure:"master"`
	Slaves []DatabaseNode `mapstructure:"slaves"`

	MaxOpenConnections int           `mapstructure:"max_open_connections"`
	MaxIdleConnections int           `mapstructure:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`
}

// DatabaseNode holds connection parameters for a single database node.
type DatabaseNode struct {
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
	User    string
	Pass    string
	Name    string `mapstructure:"name"`
	SSLMode string `mapstructure:"ssl_mode"`
}

// DSN returns the PostgreSQL DSN string for connecting to this database node.
func (n DatabaseNode) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		n.User, n.Pass, n.Host, n.Port, n.Name, n.SSLMode,
	)
}

// JWT holds JWT-related configuration.
type JWT struct {
	Secret string        `mapstructure:"secret"`
	TTL    time.Duration `mapstructure:"ttl"`
}

func MustLoad() *Config {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to read config")
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		zlog.Logger.Panic().Err(err).Msgf("failed to unmarshal config: %v", err)
	}

	cfg.Database.Master.User = os.Getenv("DB_USER")
	cfg.Database.Master.Pass = os.Getenv("DB_PASSWORD")

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	return &cfg
}
