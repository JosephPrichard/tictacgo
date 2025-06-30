package utils

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	m map[string]string
}

func NewConfig() Config {
	m := make(map[string]string)

	content, err := os.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	contentStr := string(content)

	for _, line := range strings.Split(contentStr, "\n") {
		index := strings.Index(line, "=")
		m[line[0:index]] = strings.TrimRight(line[index+1:], "\n\r\t")
	}

	return Config{m: m}
}

func (c Config) Get(key string) string {
	v, ok := c.m[key]
	if !ok {
		v, ok = os.LookupEnv(key)
		if !ok {
			log.Fatalf("error reading config key, does not exist: %s", key)
		}
		return v
	}
	return v
}
