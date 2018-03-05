package config

import (
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/gusseleet/lora-app-server/internal/common"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/gusseleet/lora-app-server/internal/nsclient"
)

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogLevel               int `mapstructure:"log_level"`
		PasswordHashIterations int `mapstructure:"password_hash_iterations"`
	}

	PostgreSQL struct {
		DSN         string `mapstructure:"dsn"`
		Automigrate bool
		DB          *common.DBLogger `mapstructure:"db"`
	} `mapstructure:"postgresql"`

	Redis struct {
		URL  string `mapstructure:"url"`
		Pool *redis.Pool
	}

	ApplicationServer struct {
		ID string `mapstructure:"id"`

		Integration struct {
			Handler handler.Handler

			MQTT struct {
				Server   string
				Username string
				Password string
				CACert   string `mapstructure:"ca_cert"`
				TLSCert  string `mapstructure:"tls_cert"`
				TLSKey   string `mapstructure:"tls_key"`
			} `mapstructure:"mqtt"`
		}

		API struct {
			Bind       string
			CACert     string `mapstructure:"ca_cert"`
			TLSCert    string `mapstructure:"tls_cert"`
			TLSKey     string `mapstructure:"tls_key"`
			PublicHost string `mapstructure:"public_host"`
		} `mapstructure:"api"`

		ExternalAPI struct {
			Bind                       string
			TLSCert                    string `mapstructure:"tls_cert"`
			TLSKey                     string `mapstructure:"tls_key"`
			JWTSecret                  string `mapstructure:"jwt_secret"`
			DisableAssignExistingUsers bool   `mapstructure:"disable_assign_existing_users"`
		} `mapstructure:"external_api"`

		Branding struct {
			Header       string
			Footer       string
			Registration string
		}

		GatewayDiscovery struct {
			Enabled   bool
			Interval  time.Duration
			Frequency int
			DR        int `mapstructure:"dr"`
		} `mapstructure:"gateway_discovery"`
	} `mapstructure:"application_server"`

	JoinServer struct {
		Bind    string
		CACert  string `mapstructure:"ca_cert"`
		TLSCert string `mapstructure:"tls_cert"`
		TLSKey  string `mapstructure:"tls_key"`
	} `mapstructure:"join_server"`

	NetworkServer struct {
		Server string
		Pool   nsclient.Pool
	} `mapstructure:"network_server"`
}

// C holds the global configuration.
var C Config
