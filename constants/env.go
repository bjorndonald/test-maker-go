package constants

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string

	DbHost         string
	DbUser         string
	DbPassword     string
	DbName         string
	DbPort         string
	SSLMode        string
	OpenAIKey      string
	EmbeddingModel string
}

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Printf("error loading .env file %s", err)
	}

}

func New() *Config {

	log.Println("app port env =>", getEnv("PORT", "8000"))

	return &Config{
		Port: getEnv("PORT", "8000"),
		Env:  getEnv("ENV", "development"),

		DbHost:     getEnv("POSTGRES_HOST", ""),
		DbUser:     getEnv("POSTGRES_USER", ""),
		DbPassword: getEnv("POSTGRES_PASSWORD", ""),
		DbName:     getEnv("POSTGRES_NAME", ""),
		DbPort:     getEnv("POSTGRES_PORT", ""),
		SSLMode:    getEnv("SSL_MODE", "disable"),

		OpenAIKey:      getEnv("OPENAI_API_KEY", ""),
		EmbeddingModel: getEnv("EMBEDDING_MODEL", ""),
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
