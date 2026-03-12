package app

import (
	"encoding/json"
	"os"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

type App struct {
	Port          Port          `json:"port"`
	RedisURL      string        `json:"env:redis-url.key"`
	OpenAIKey     string        `json:"env:open-ai.key"`
	RateLimits    RateLimits    `json:"rate-limit:limits"`
	ClientOptions ClientOptions `json:"redis:client-opt"`
}

type Port struct {
	Default string `json:"default"`
	Env     string `json:"env"`
}

type RateLimits struct {
	PerMinute int `json:"RequestPerMinute"`
	PerDay    int `json:"RequestPerDay"`
}

type ClientOptions struct {
	DialTimeout  Duration `json:"dialTimeout"`
	ReadTimeout  Duration `json:"readTimeout"`
	WriteTimeout Duration `json:"writeTimeout"`
	PoolTimeout  Duration `json:"poolTimeout"`
	MaxRetries   int      `json:"maxRetries"`
}

func NewApplication() (*App, error) {
	app, err := loadJSON("./application_settings.json")
	if err != nil {
		return nil, err
	}

	app.OpenAIKey = os.Getenv(app.OpenAIKey)
	app.RedisURL = os.Getenv(app.RedisURL)

	port := os.Getenv(app.Port.Env)
	if port == "" {
		port = app.Port.Default
	}
	app.Port.Env = port

	return app, nil
}

func loadJSON(path string) (*App, error) {
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
