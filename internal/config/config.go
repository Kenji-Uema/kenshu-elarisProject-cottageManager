package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type MongoDbConfig struct {
	Url      string `env:"MONGO_URL,required"`
	Port     string `env:"MONGO_PORT,required"`
	Db       string `env:"MONGO_DB,required"`
	User     string `env:"MONGO_USER,required"`
	Password string `env:"MONGO_PASSWORD,required"`
}

type CottageCollectionConfig struct {
	Name string `env:"COTTAGE_COLLECTION" envDefault:"Cottage"`
}

type BookingCollectionConfig struct {
	Name string `env:"BOOKING_COLLECTION" envDefault:"Booking"`
}

type GinConfig struct {
	Port string `env:"GIN_PORT" envDefault:"8080"`
}

type LogConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

type TelemetryConfig struct {
	ExporterEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	ServiceName      string `env:"OTEL_SERVICE_NAME" envDefault:"cottage-manager"`
	DeploymentEnv    string `env:"DEPLOYMENT_ENV" envDefault:"development"`
	ServiceVersion   string `env:"SERVICE_VERSION" envDefault:"0.0.1"`
	UseInsecure      bool   `env:"OTEL_EXPORTER_OTLP_INSECURE" envDefault:"true"`
}

func LoadConfig[C GinConfig | MongoDbConfig |
	CottageCollectionConfig | BookingCollectionConfig |
	LogConfig | TelemetryConfig]() *C {
	var c C
	if err := env.Parse(&c); err != nil {
		log.Fatal(err)
	}

	return &c
}
