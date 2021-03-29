package config

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfigCreatedViaEnv(t *testing.T) {
	osConfigValue := map[string]string{
		"SERVER_HOST": "localhost:2000",
		"HIDE_IP":     "true",
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

// TODO implement this
func TestConfigCreatedViaFile(t *testing.T) {
}
