package config

import (
	"encoding/base64"
	"os"
	"testing"
	"time"

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
		Timeout:  time.Second * 30,
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

	if !cfg.CheckAuthString("Basic " + authString) {
		t.Errorf("expected true for verified, got false")
	}
}

func TestALlowedIP(t *testing.T) {
	t.Run("AllowAll", func(t *testing.T) {
		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("error creating config: %v", err)
		}

		if got := cfg.ShouldAllowIP("127.0.0.1:123"); !got {
			t.Errorf("expected true got %t", got)
		}

	})

	t.Run("NoAllowAll", func(t *testing.T) {
		if err := os.Setenv("ALLOWEDIP", "127.0.0.1"); err != nil {
			t.Fatalf("error setting env value %v", err)
		}

		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("error creating config: %v", err)
		}

		testCases := []struct {
			address  string
			expected bool
		}{
			{
				address:  "127.0.0.1:123",
				expected: true,
			},
			{
				address:  "127.0.0.1",
				expected: true,
			},
			{
				address:  "127.0.0.3",
				expected: false,
			},
		}

		for _, item := range testCases {
			if got := cfg.ShouldAllowIP(item.address); got != item.expected {
				t.Errorf("expected %t for %s , got %t", item.expected, item.address, got)
			}
		}

	})
}
