package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Conf ...
var Conf *Config

// Config ...
type Config struct {
	Port string
	// MessagePath string
	// MaxWorker   int
	// MaxQueue    int
	// MaxLength   int
	DatabaseURL string
	//PubSubConfig
	//Pusher
	SentryUrlDSN  string
	SecretKeyJWT  string
	LoggerDefault string
	LogLevel      int
}

// New returns a new Config struct
func New() *Config {

	if tz := getEnv("TIMEZONE", "Local"); tz != "" {
		var err error
		time.Local, err = time.LoadLocation(tz)
		if err != nil {
			log.Printf("[ERROR] loading location '%s': %v\n", tz, err)
		}
		nameLocation, _ := time.Now().Zone()
		log.Printf("[INFO] текущая таймзона %s", nameLocation)
	}

	c := &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DB_CONNECTION", "postgres://postgres:@localhost/go-keeper?sslmode=disable"),
		SentryUrlDSN:  getEnv("SENTRY_URL_DSN", "http://ded6d3a6b5c64c0d9e38c042a365fa39:0aed9c2ef0994bf39f40e7227174bfa2@localhost:9000/2"),
		SecretKeyJWT:  getEnv("SECRET_KEY_JWT", ""),
		LoggerDefault: getEnv("LOGGER_DEFAULT", "Sentry"),
	}

	Conf = c
	return c
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Helper to read an environment variable into a string slice or return default value
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
