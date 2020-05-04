package configuration

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	flag "github.com/spf13/pflag"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	flagForDBUser         = "db_user"
	flagForDBPass         = "db_pass"
	flagForDBHost         = "db_host"
	flagForDBName         = "db_name"
	flagForGrpcPort       = "grpc_port"
	flagForGrpcPortHealth = "grpc_port_health"
	flagForHTTPPort       = "http_port"
)

type Config struct {
	// AppName is a low cardinality identifier for this service.
	AppName string

	// AppID is a unique identifier for the instance (pod) running this app.
	AppID string

	DBHost string
	DBUser string
	DBPass string
	DBName string

	// GrpcPort controls what port our gRPC server runs on.
	GrpcPort int

	// HTTPPort controls what port our HTTP server runs on.
	HTTPPort int

	// HealthPort controls what port our gRPC health endpoints run on.
	HealthPort int
}

func (c Config) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		log.Fatalf("Could not marshal config to string: %v", err)
	}
	return string(b)
}

func LoadConfig() Config {
	c := Config{
		AppName:    "password-reset",
		AppID:      uuid.New().String(),
		GrpcPort:   8081,
		HealthPort: 8082,
		HTTPPort:   8083,
		DBName:     "postgres",
		DBUser:     "postgres",
		DBPass:     "postgres",
	}
	c.DBHost = fmt.Sprintf("%s-db", c.DBName)

	flag.String(flagForDBUser, c.DBUser, "DB user")
	flag.String(flagForDBPass, c.DBPass, "DB pass")
	flag.String(flagForDBHost, c.DBHost, "DB host")
	flag.String(flagForDBName, c.DBName, "DB name")

	flag.Int(flagForGrpcPort, c.GrpcPort, "gRPC port")
	flag.Int(flagForGrpcPortHealth, c.HealthPort, "gRPC health port")
	flag.Int(flagForHTTPPort, c.HTTPPort, "gRPC HTTP port")

	flag.Parse()

	viper.BindPFlag(flagForDBUser, flag.Lookup(flagForDBUser))
	viper.BindPFlag(flagForDBPass, flag.Lookup(flagForDBPass))
	viper.BindPFlag(flagForDBHost, flag.Lookup(flagForDBHost))
	viper.BindPFlag(flagForDBName, flag.Lookup(flagForDBName))

	viper.BindPFlag(flagForGrpcPort, flag.Lookup(flagForGrpcPort))
	viper.BindPFlag(flagForGrpcPortHealth, flag.Lookup(flagForGrpcPortHealth))
	viper.BindPFlag(flagForHTTPPort, flag.Lookup(flagForHTTPPort))

	viper.AutomaticEnv()

	c.DBUser = viper.GetString(flagForDBUser)
	c.DBPass = viper.GetString(flagForDBPass)
	c.DBHost = viper.GetString(flagForDBHost)
	c.DBName = viper.GetString(flagForDBName)

	c.GrpcPort = viper.GetInt(flagForGrpcPort)
	c.HealthPort = viper.GetInt(flagForGrpcPortHealth)
	c.HTTPPort = viper.GetInt(flagForHTTPPort)

	return c
}

func getGrpcAddress(addrFlag, hostFlag, portFlag string) string {
	addr := viper.GetString(addrFlag)
	host := viper.GetString(hostFlag)
	port := viper.GetInt(portFlag)

	if port != 0 {
		return fmt.Sprintf("%s:%d", host, port)
	}

	return addr
}
