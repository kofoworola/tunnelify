package config

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfigCreatedViaEnv(t *testing.T) {
	osConfigValue := map[string]string{
		"SERVER_HOST": "localhost:2000",
		"HIDEIP":      "true",
	}
	for key, val := range osConfigValue {
		if err := os.Setenv(key, val); err != nil {
			t.Fatalf("error setting env value %s", key)
		}
	}
	want := &Config{
		HostName: osConfigValue["SERVER_HOST"],
		HideIP:   true,
	}

	got, err := LoadConfig("")
	if err != nil {
		t.Fatalf("error creating config: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("LoadConfig mismatch (-want,+got):\n%s", diff)
	}
}

func TestConfigAuthCheck(t *testing.T) {
	authString := base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if err := os.Setenv("SERVER_AUTH", "user:pass"); err != nil {
		t.Fatalf("error setting env value %v", err)
	}
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("error creating config: %v", err)
	}

	if !cfg.HasAuth() {
		t.Errorf("expected true for HasAuth, got false instead")
	}

	verified, err := cfg.CheckAuthString("Basic " + authString)
	if err != nil {
		t.Fatalf("error verifying auth string: %v", err)
	}

	if !verified {
		t.Errorf("expected true for verified, got false")
	}
}
