package config

import (
	"log"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type AppConfig struct {
	AppEnv  string `mapstructure:"APP_ENV" validate:"required,oneof=development production testing"`
	Port    int    `mapstructure:"PORT" validate:"required"`
	AppName string `mapstructure:"APP_NAME"`

	DBHost     string `mapstructure:"DB_HOST" validate:"required"`
	DBUser     string `mapstructure:"DB_USER" validate:"required"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME" validate:"required"`
	DBPort     int    `mapstructure:"DB_PORT" validate:"required"`
	DBSSLMode  string `mapstructure:"DB_SSLMODE" validate:"required"`

	// ... (Tambahkan field lain sesuai kebutuhanmu: Redis, Kafka, JWT, dll)
	CorsAllowedOrigins string        `mapstructure:"CORS_ALLOWED_ORIGINS" validate:"required"`
	RateLimitMax       int           `mapstructure:"RATE_LIMIT_MAX" validate:"required"`
	RateLimitExp       time.Duration `mapstructure:"RATE_LIMIT_EXP" validate:"required"`

	// Contoh Auth
	ClientID     string `mapstructure:"CLIENT_ID" validate:"required"`
	ClientSecret string `mapstructure:"CLIENT_SECRET" validate:"required"`
	CallbackURL  string `mapstructure:"CALLBACK_URL" validate:"required"`
}

var Env AppConfig

func LoadConfig() {
	// 1. Cari file .env (Untuk development pakai 'air' di laptop)
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..") // Berjaga-jaga kalau run dari cmd/ folder

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Warning: .env file not found. Reading entirely from OS Environment Variables...")
	}

	// 2. Baca dari OS Environment (Untuk Docker)
	viper.AutomaticEnv()

	// 3. 🚨 THE FIX: Otomatis daftarkan semua tag 'mapstructure' agar dibaca dari OS ENV
	bindStructEnvs(AppConfig{})

	// 4. Masukkan ke Struct
	if err := viper.Unmarshal(&Env); err != nil {
		log.Fatalf("Environment can't be loaded: %v", err)
	}

	// 5. Validasi Struct
	validate := validator.New()
	if err := validate.Struct(&Env); err != nil {
		log.Fatalf("Environment configuration validation failed: %v", err)
	}

	log.Println("✅ Configuration loaded successfully!")
}

// bindStructEnvs adalah fungsi helper untuk mengatasi masalah Viper di Docker
func bindStructEnvs(v interface{}) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag != "" {
			// Kasih tau Viper: "Tolong cari environment variable dengan nama ini"
			viper.BindEnv(tag)
		}
	}
}
