package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName              string
	AppEnv               string
	HTTPPort             string
	DatabaseURL          string
	JWTSecret            string
	AccessTokenExpireMin int

	// W6 is configuration for the IECube W6 AI gateway.
	// All fields are loaded from environment variables (see .env.example).
	W6 W6Config
}

// W6Config holds settings for the third-party W6 AI service.
type W6Config struct {
	BaseURL       string
	WSSBaseURL    string
	AuthHeaderKey string
	AuthHeaderVal string
	ModelProcedure string
	ModelLLM       string
	ModelLLMShort  string
	ModuleName     string
}

// Load 从 .env 加载配置，所有配置以 .env 为准，无代码内默认值（除端口为空时用 8080 以便启动）
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env not found: %v", err)
	}
	cfg := &Config{
		AppName:              getEnv("APP_NAME"),
		AppEnv:               getEnv("APP_ENV"),
		HTTPPort:             getEnv("HTTP_PORT"),
		DatabaseURL:          getEnv("DATABASE_URL"),
		JWTSecret:            getEnv("JWT_SECRET"),
		AccessTokenExpireMin: getEnvInt("ACCESS_TOKEN_EXPIRE_MINUTES"),
		W6: W6Config{
			BaseURL:        getEnv("W6_BASE_URL"),
			WSSBaseURL:     getEnv("W6_WSS_BASE_URL"),
			AuthHeaderKey:  getEnv("W6_AUTH_HEADER_FIELD"),
			AuthHeaderVal:  getEnv("W6_AUTH_HEADER_VALUE"),
			ModelProcedure: getEnv("W6_MODEL_PROCEDURE"),
			ModelLLM:       getEnv("W6_MODEL_LLM"),
			ModelLLMShort:  getEnv("W6_MODEL_LLM_SHORT"),
			ModuleName:     getEnv("W6_MODULE_NAME"),
		},
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required (set in .env)")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required (set in .env)")
	}
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8080"
	}
	if cfg.AppName == "" {
		cfg.AppName = "YouMind Backend v2"
	}
	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}
	if cfg.AccessTokenExpireMin == 0 {
		cfg.AccessTokenExpireMin = 60
	}

	// W6 configuration is optional at startup; but if BaseURL is set, we require
	// the essential auth fields to be present to avoid confusing runtime errors.
	if cfg.W6.BaseURL != "" {
		if cfg.W6.AuthHeaderKey == "" || cfg.W6.AuthHeaderVal == "" {
			log.Fatal("W6_BASE_URL is set but W6_AUTH_HEADER_FIELD or W6_AUTH_HEADER_VALUE is missing in .env")
		}
		if cfg.W6.WSSBaseURL == "" {
			log.Fatal("W6_WSS_BASE_URL is required when using W6 AI")
		}
		if cfg.W6.ModelProcedure == "" {
			cfg.W6.ModelProcedure = "raw"
		}
	}
	return cfg
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func getEnvInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}
