package bootstrap

import (
	"encoding/json"
	"os"
	"time"
)

type App struct {
	Port          string        `json:"env:port.key"`
	RedisURL      string        `json:"env:redis-url.key"`
	OpenAIKey     string        `json:"env:open-ai.key"`
	RateLimits    RateLimits    `json:"rate-limit:limits"`
	ClientOptions ClientOptions `json:"redis:client-opt"`
}
type RateLimits struct {
	PerMinute int `json:"RequestPerMinute"`
	PerDay    int `json:"RequestPerDay"`
}
type ClientOptions struct {
	DialTimeout  time.Duration `json:"dialTimeout"`
	ReadTimeout  time.Duration `json:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout"`
	PoolTimeout  time.Duration `json:"poolTimeout"`
	MaxRetries   int           `json:"maxRetries"`
}

func getAppSettingsJson(path string) (*App, error) {
	// reader of app json path
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var app App
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&app); err != nil {
		return nil, err
	}
	return &app, nil
}
