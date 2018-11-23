package config

import (
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/handler/gcppubsub"
	"github.com/brocaar/lora-app-server/internal/handler/mqtthandler"
	"github.com/brocaar/lora-app-server/internal/nsclient"
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
		URL         string        `mapstructure:"url"`
		MaxIdle     int           `mapstructure:"max_idle"`
		IdleTimeout time.Duration `mapstructure:"idle_timeout"`
		Pool        *redis.Pool
	}

	ApplicationServer struct {
		ID string `mapstructure:"id"`

		Integration struct {
			Backend   string
			Handler   handler.Handler
			MQTT      mqtthandler.Config `mapstructure:"mqtt"`
			GCPPubSub gcppubsub.Config   `mapstructure:"gcp_pub_sub"`
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
			CORS                       struct {
				AllowOrigin string `mapstructure:"allow_origin"`
			} `mapstructure:"cors"`
		} `mapstructure:"external_api"`

		Branding struct {
			Header       string
			Footer       string
			Registration string
		}
	} `mapstructure:"application_server"`

	JoinServer struct {
		Bind    string
		CACert  string `mapstructure:"ca_cert"`
		TLSCert string `mapstructure:"tls_cert"`
		TLSKey  string `mapstructure:"tls_key"`

		KEK struct {
			ASKEKLabel string `mapstructure:"as_kek_label"`

			Set []struct {
				Label string `mapstructure:"label"`
				KEK   string `mapstructure:"kek"`
			}
		} `mapstructure:"kek"`
	} `mapstructure:"join_server"`

	NetworkServer struct {
		Pool nsclient.Pool
	} `mapstructure:"network_server"`
}

// C holds the global configuration.
var C Config
