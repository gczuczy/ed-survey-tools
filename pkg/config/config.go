package config

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	DB       DBConfig       `koanf:"db"`
	HTTP     HTTPConfig     `koanf:"http"`
	OAuth2   OAuth2Config   `koanf:"oauth2"`
	Sessions SessionsConfig `koanf:"sessions"`
	Logging  LoggingConfig  `koanf:"logging"`
	VSDS     VSDSConfig     `koanf:"vsds"`
	EDSM     EDSMConfig     `koanf:"edsm"`
	Bundles  BundlesConfig  `koanf:"bundles"`
}

type BundlesConfig struct {
	Path          string        `koanf:"path"`
	Serve         bool          `koanf:"serve"`
	BaseURL       string        `koanf:"baseUrl"`
	CheckInterval time.Duration `koanf:"checkInterval"`
}

type EDSMConfig struct {
	Timeout time.Duration `koanf:"timeout"`
	Retries int           `koanf:"retries"`
}

type VSDSConfig struct {
	ProcessorInterval time.Duration `koanf:"processorInterval"`
}

type DBConfig struct {
	Host     string  `koanf:"host"`
	Port     *uint16 `koanf:"port"`
	Database string  `koanf:"dbname"`
	User     string  `koanf:"user"`
	Password string  `koanf:"password"`
	MaxConns int32   `koanf:"maxconns"`
	MinConns int32   `koanf:"minconns"`
	SSL      bool    `koanf:"ssl"`
}

type HTTPConfig struct {
	Port uint16 `koanf:"port"`
}

type OAuth2Config struct {
	ClientID string `koanf:"clientid"`
	ClientSecret string `koanf:"clientsecret"`
	Issuer string `koanf:"issuer"`
	AuthorizeURL string `koanf:"authorizeUrl"`
	TokenURL string `koanf:"tokenUrl"`
	UserInfoURL string `koanf:"userinfoUrl"`
	ExtraScopes []string `koanf:"extrascopes"`
}

type SessionsConfig struct {
	Key    string              `koanf:"key"`
	Store  string              `koanf:"store"`
	Secure bool                `koanf:"secure"`
	MaxAge int                 `koanf:"maxAge"`
	Redis  *RedisSessionConfig `koanf:"redis"`
}

type RedisSessionConfig struct {
	MaxIdle *int `koanf:"maxidle"`
	IdleTimeout *time.Duration `koanf:"idletimeout"`
	DB *int `koanf:"db"`
	User *string `koanf:"user"`
	Pass *string `koanf:"pass"`
	Host *string `koanf:"host"`
	Port *uint16 `koanf:"port"`
}

type SyslogConfig struct {
	Host     string `koanf:"host"`
	Port     uint16 `koanf:"port"`
	Proto    string `koanf:"proto"`
	Facility string `koanf:"facility"`
}

type LoggingConfig struct {
	Level     string       `koanf:"level"`
	Output    string       `koanf:"output"`
	Timestamp bool         `koanf:"timestamp"`
	Syslog    SyslogConfig `koanf:"syslog"`
}

func ParseConfig(k *koanf.Koanf) (*Config, error) {
	var (
		err error
	)
	cfgfile := k.String(`config`)

	if err = k.Load(file.Provider(cfgfile), yaml.Parser()); err != nil {
		return nil, err
	}

	cfg := Config{
		DB: DBConfig{
			MaxConns: 8,
			MinConns: 1,
			SSL:      false,
		},
		HTTP: HTTPConfig{
			Port: 80,
		},
		OAuth2: OAuth2Config{
			Issuer: "https://auth.frontierstore.net",
		},
		Logging: LoggingConfig{
			Level:     "info",
			Output:    "stdio",
			Timestamp: false,
			Syslog: SyslogConfig{
				Host:     "127.0.0.1",
				Port:     514,
				Facility: "LOCAL0",
			},
		},
		VSDS: VSDSConfig{
			ProcessorInterval: time.Minute,
		},
		EDSM: EDSMConfig{
			Timeout: 5 * time.Second,
			Retries: 10,
		},
		Bundles: BundlesConfig{
			CheckInterval: 5 * time.Minute,
		},
		Sessions: SessionsConfig{
			MaxAge: 7200,
		},
	}
	if err = k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	// adjust oauth2 config
	if len(cfg.OAuth2.AuthorizeURL)==0 {
		cfg.OAuth2.AuthorizeURL = fmt.Sprintf("%s/auth", cfg.OAuth2.Issuer)
	}
	if len(cfg.OAuth2.TokenURL)==0 {
		cfg.OAuth2.TokenURL = fmt.Sprintf("%s/token", cfg.OAuth2.Issuer)
	}
	if len(cfg.OAuth2.UserInfoURL)==0 {
		cfg.OAuth2.UserInfoURL = fmt.Sprintf("%s/decode", cfg.OAuth2.Issuer)
	}

	return &cfg, nil
}
