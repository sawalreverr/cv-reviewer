package config

import (
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Storage  StorageConfig
    Gemini GeminiConfig
    Queue QueueConfig
}

type ServerConfig struct {
    Port string
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type StorageConfig struct {
    UploadDir string
}

type GeminiConfig struct {
    APIKey string
    Model string
    EmbeddingModel string
    Temperature float32
    MaxTokens int32
    Dimension *int32
}

type QueueConfig struct {
    WorkerCount int
    QueueSize int
    JobTimeout int
}

func Load() (*Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    config := &Config{
        Server: ServerConfig{
            Port: viper.GetString("SERVER_PORT"),
        },
        Database: DatabaseConfig{
            Host:     viper.GetString("DB_HOST"),
            Port:     viper.GetString("DB_PORT"),
            User:     viper.GetString("DB_USER"),
            Password: viper.GetString("DB_PASSWORD"),
            DBName:   viper.GetString("DB_NAME"),
            SSLMode:  viper.GetString("DB_SSLMODE"),
        },
        Storage: StorageConfig{
            UploadDir: viper.GetString("UPLOAD_DIR"),
        },
        Gemini: GeminiConfig{
            APIKey: viper.GetString("GEMINI_APIKEY"),
            Model: viper.GetString("GEMINI_MODEL"),
            EmbeddingModel: viper.GetString("GEMINI_EMBEDDING_MODEL"),
            Temperature: float32(viper.GetFloat64("GEMINI_TEMPERATURE")),
			MaxTokens: viper.GetInt32("GEMINI_MAX_TOKENS"),
            Dimension: parseDimensionPtr(),
        },
        Queue: QueueConfig {
        	WorkerCount: viper.GetInt("WORKER_COUNT"),
        	QueueSize: viper.GetInt("JOB_QUEUE_SIZE"),
        	JobTimeout: viper.GetInt("JOB_TIMEOUT"),
        },
    }
    
    return config, nil
}

func parseDimensionPtr() *int32 {
    dimension := viper.GetString("GEMINI_DIMENSION")

    val, _ := strconv.Atoi(dimension)
    val32 := int32(val)

    return &val32
}