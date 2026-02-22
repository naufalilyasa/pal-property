package config

import (
	"log"
	"reflect"
	"strings"
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

	// Redis
	RedisAddr          string        `mapstructure:"REDIS_ADDR" validate:"required"`
	RedisPassword      string        `mapstructure:"REDIS_PASSWORD"`
	RedisDB            int           `mapstructure:"REDIS_DB"`
	CorsAllowedOrigins string        `mapstructure:"CORS_ALLOWED_ORIGINS" validate:"required"`
	RateLimitMax       int           `mapstructure:"RATE_LIMIT_MAX" validate:"required"`
	RateLimitExp       time.Duration `mapstructure:"RATE_LIMIT_EXP" validate:"required"`

	// Contoh Auth
	ClientID     string `mapstructure:"CLIENT_ID" validate:"required"`
	ClientSecret string `mapstructure:"CLIENT_SECRET" validate:"required"`
	CallbackURL  string `mapstructure:"CALLBACK_URL" validate:"required"`

	// JWT
	JwtPrivateKeyBase64  string        `mapstructure:"JWT_PRIVATE_KEY_BASE64" validate:"required"`
	JwtPublicKeyBase64   string        `mapstructure:"JWT_PUBLIC_KEY_BASE64" validate:"required"`
	JwtAccessExpiration  time.Duration `mapstructure:"JWT_ACCESS_EXPIRATION" validate:"required"`
	JwtRefreshExpiration time.Duration `mapstructure:"JWT_REFRESH_EXPIRATION" validate:"required"`
}

var Env AppConfig

func LoadConfig() {
	// 1. Cari file .env (Untuk development pakai 'air' di laptop)
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..") // Berjaga-jaga kalau run dari cmd/ folder

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Warning: .env file not found. Reading entirely from OS Environment Variables...")
	}

	// 2. Baca dari OS Environment (Untuk Docker)
	viper.AutomaticEnv()

	// 3. Otomatis daftarkan semua tag 'mapstructure' agar dibaca dari OS ENV
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
