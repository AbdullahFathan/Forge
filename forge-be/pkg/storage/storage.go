package storage

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func New(cfg Config) Config {
	return cfg
}
