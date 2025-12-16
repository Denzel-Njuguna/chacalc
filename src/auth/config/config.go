package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SupabaseURL     string
	SupabaseAnonKey string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		println("Env error")
	}

	cfg := &Config{
		SupabaseURL:     os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey: os.Getenv("SUPABASE_ANON_KEY"),
	}

	if cfg.SupabaseURL == "" || cfg.SupabaseAnonKey == "" {
		log.Fatal("Missing Supabase ENV")
	}

	return cfg
}
