package config

import (
	"errors"

	"github.com/caarlos0/env/v11"
)

type Configs struct {
	AppConfig
	MongoConfig
	CottageCollectionConfig
	BookingCollectionConfig
	LogConfig
	TelemetryConfig
}

type AppConfig struct {
	ServiceName string `env:"SERVICE_NAME"`
	Version     string `env:"VERSION"`
	Host        string `env:"SERVICE_HOST,required"`
	Port        int    `env:"SERVICE_PORT,required"`
}

type MongoConfig struct {
	Username string `env:"MONGO_INITDB_ROOT_USERNAME,required"`
	Password string `env:"MONGO_INITDB_ROOT_PASSWORD,required"`
	Host     string `env:"MONGO_HOST,required"`
	Database string `env:"MONGO_DATABASE,required"`
}

type CottageCollectionConfig struct {
	Name string `env:"COTTAGE_COLLECTION" envDefault:"Cottage"`
}

type BookingCollectionConfig struct {
	Name string `env:"BOOKING_COLLECTION" envDefault:"Booking"`
}

type LogConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

type TelemetryConfig struct {
	OTLPEndpoint   string `env:"OTEL_EXPORTER_OTLP_ENDPOINT,required"`
	OTLPGrpcPort   int    `env:"OTEL_EXPORTER_OTLP_GRPC_PORT,required"`
	OTLPHealthPort int    `env:"OTEL_EXPORTER_OTLP_HEALTH_PORT,required"`
	OTLPInsecure   bool   `env:"OTEL_EXPORTER_OTLP_INSECURE,required"`
}

func LoadConfigs() (Configs, error) {
	var err error

	appConfig, loadErr := loadConfig[AppConfig]()
	err = errors.Join(err, loadErr)
	mongoConfig, loadErr := loadConfig[MongoConfig]()
	err = errors.Join(err, loadErr)
	cottageCollectionConfig, loadErr := loadConfig[CottageCollectionConfig]()
	err = errors.Join(err, loadErr)
	bookingCollectionConfig, loadErr := loadConfig[BookingCollectionConfig]()
	err = errors.Join(err, loadErr)
	logCofig, loadErr := loadConfig[LogConfig]()
	err = errors.Join(err, loadErr)
	telemetryConfig, loadErr := loadConfig[TelemetryConfig]()
	err = errors.Join(err, loadErr)

	return Configs{
		appConfig,
		mongoConfig,
		cottageCollectionConfig,
		bookingCollectionConfig,
		logCofig,
		telemetryConfig,
	}, err
}

func loadConfig[C AppConfig | MongoConfig | CottageCollectionConfig | BookingCollectionConfig | LogConfig | TelemetryConfig]() (C, error) {
	var c C
	if err := env.Parse(&c); err != nil {
		return c, err
	}

	return c, nil
}
