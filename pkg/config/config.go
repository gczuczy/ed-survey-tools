package config

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	DB DBConfig `koanf:"db"`
	HTTP HTTPConfig `koanf:"http"`
	OAuth2 OAuth2Config `koanf:"oauth2"`
}

type DBConfig struct {
	Host string `koanf:"host"`
	Port *uint16 `koanf:"port"`
	Database string `koanf:"dbname"`
	User string `koanf:"user"`
	Password string `koanf:"password"`
	MaxConns int32 `koanf:"maxconns"`
	MinConns int32 `koanf:"minconns"`
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
		},
		HTTP: HTTPConfig{
			Port: 80,
		},
		OAuth2: OAuth2Config{
			Issuer: "https://auth.frontierstore.net",
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
