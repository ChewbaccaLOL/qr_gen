package cli

import (
	"bufio"
	"os"
	"strings"
)

type Env interface {
	Lookup(key string) (string, bool)
	Set(key, value string)
}

type OSEnv struct{}

func (OSEnv) Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (OSEnv) Set(key, value string) {
	if _, exists := os.LookupEnv(key); exists {
		return
	}
	_ = os.Setenv(key, value)
}

type MapEnv struct {
	Values map[string]string
}

func (m *MapEnv) Lookup(key string) (string, bool) {
	if m == nil || m.Values == nil {
		return "", false
	}
	value, ok := m.Values[key]
	return value, ok
}

func (m *MapEnv) Set(key, value string) {
	if m == nil {
		return
	}
	if m.Values == nil {
		m.Values = map[string]string{}
	}
	if _, exists := m.Values[key]; exists {
		return
	}
	m.Values[key] = value
}

func LoadDotenv(path string, env Env) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key == "" {
			continue
		}
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		env.Set(key, value)
	}
}
