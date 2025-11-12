package config

import "github.com/spf13/viper"

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Storage  StorageConfig
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
    }
    
    return config, nil
}