package config

import "os"

func Mode() string {
	if v := os.Getenv("APP_MODE"); v != "" {
		return v
	}
	return "normal"
}
