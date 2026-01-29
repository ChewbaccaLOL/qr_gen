package gui

import (
	"os"
	"strings"
)

func IsWSL() bool {
	return detectWSL(os.Environ(), readProcVersion())
}

func detectWSL(envEntries []string, procVersion string) bool {
	if hasEnvValue(envEntries, "WSL_INTEROP") ||
		hasEnvValue(envEntries, "WSL_DISTRO_NAME") ||
		hasEnvValue(envEntries, "WSLENV") {
		return true
	}
	version := strings.ToLower(procVersion)
	return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
}

func hasEnvValue(envEntries []string, key string) bool {
	prefix := key + "="
	for _, entry := range envEntries {
		if strings.HasPrefix(entry, prefix) {
			return true
		}
	}
	return false
}

func readProcVersion() string {
	if data, err := os.ReadFile("/proc/version"); err == nil {
		return string(data)
	}
	if data, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		return string(data)
	}
	return ""
}
