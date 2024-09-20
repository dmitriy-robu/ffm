package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"go-fitness/external/logger/sl"
	"log/slog"
	"os"
	"time"
)

type (
	Config struct {
		Env        string `yaml:"env"`
		HTTPServer `yaml:"http_server"`
		WSServer   `yaml:"ws_server"`
		ENVState   `yaml:"env_state"`
		DB         `yaml:"db"`
		Video      `yaml:"video_service"`
		JWT        string `yaml:"jwt_secret" env:"JWT_SECRET"`
	}

	DB struct {
		MaxOpenConns    int           `yaml:"max_open_conns" env:"MAX_OPEN_CONNS" env-default:"25"`
		MaxIdleConns    int           `yaml:"max_idle_conns"  env:"MAX_IDLE_CONNS" env-default:"25"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"MAX_lifetime_CONNS" env-default:"3m"`
		MysqlUser       string        `env:"MYSQL_USER" env-default:"root"`
		MysqlPassword   string        `env:"MYSQL_PASSWORD" env-default:"root"`
		MysqlHost       string        `env:"MYSQL_HOST" env-default:"localhost"`
		MysqlPort       string        `env:"MYSQL_PORT" env-default:"3306"`
		MysqlDBName     string        `env:"MYSQL_DBNAME" env-default:"rust"`
	}

	ENVState struct {
		Local string `yaml:"local" env-default:"local"`
		Dev   string `yaml:"dev" env-default:"dev"`
		Prod  string `yaml:"prod" env-default:"prod"`
	}

	WSServer struct {
		AppID   string `yaml:"app_id" env:"PUSHER_APP_ID"`
		Host    string `yaml:"address" env:"PUSHER_HOST" env-default:"localhost"`
		Port    string `yaml:"port" env:"PUSHER_PORT" env-default:"8080"`
		Cluster string `yaml:"cluster" env:"PUSHER_CLUSTER" env-default:"ap1"`
		Secret  string `yaml:"secret" env:"PUSHER_SECRET"`
		Key     string `yaml:"key" env:"PUSHER_KEY"`
		Secure  bool   `yaml:"secure" env:"PUSHER_SECURE" env-default:"false"`
	}

	HTTPServer struct {
		ApiPort     string        `yaml:"api_port" env:"API_PORT" env-default:":8082"`
		Timeout     time.Duration `yaml:"timeout" env-default:"60s"`
		IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
		StoragePath string        `yaml:"storage_path" env:"STORAGE_PATH" env-default:"./storage"`
	}

	Video struct {
		VideoPath                 string   `yaml:"video_path" env:"VIDEO_PATH" env-default:"videos"`
		TranscodeVideoWorkerCount int      `yaml:"transcode_worker_count" env:"TRANSCODE_WORKER_COUNT" env-default:"1"`
		Resolutions               []string `yaml:"resolutions" env:"RESOLUTIONS" env-default:"360,480,720,1080"`

		//Resolutions []string `yaml:"resolutions" env:"RESOLUTIONS" env-default:"360"`
	}
)

func NewConfig() (*Config, error) {
	const op string = "config.NewConfig"

	log := slog.With(
		sl.String("op", op),
	)

	if err := godotenv.Load(".env"); err != nil {
		log.Error("error loading .env file", sl.Err(err))
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	configPath := os.Getenv("CONFIG_PATH")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Error("config file does not exist", sl.String("path", configPath))
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Error("cannot read config", sl.Err(err))
		return nil, fmt.Errorf("cannot read config: %s", err)
	}

	return &cfg, nil
}
