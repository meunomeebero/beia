package bootstrap

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
	DialTimeout  int `json:"dialTimeout"`
	ReadTimeout  int `json:"readTimeout"`
	WriteTimeout int `json:"writeTimeout"`
	PoolTimeout  int `json:"poolTimeout"`
	MaxRetries   int `json:"maxRetries"`
}
