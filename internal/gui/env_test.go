package gui

import "testing"

func TestDetectWSLFromEnv(t *testing.T) {
	env := []string{"WSL_DISTRO_NAME=Ubuntu"}
	if !detectWSL(env, "") {
		t.Fatalf("expected WSL detection from env")
	}
}

func TestDetectWSLFromVersion(t *testing.T) {
	version := "Linux version 5.15.90.1-microsoft-standard-WSL2"
	if !detectWSL(nil, version) {
		t.Fatalf("expected WSL detection from version")
	}
}

func TestDetectWSLFalse(t *testing.T) {
	version := "Linux version 6.5.0-21-generic"
	if detectWSL([]string{"PATH=/usr/bin"}, version) {
		t.Fatalf("expected non-WSL detection")
	}
}
