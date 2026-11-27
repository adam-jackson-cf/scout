package config

import "os"

func Mode() string {
	if os.Getenv("APP_TEST_MODE") != "" {
		return "fake"
	}
	return "normal"
}
