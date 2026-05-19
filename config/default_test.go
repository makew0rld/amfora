package config

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
)

// TestShowOpenedModalDefault verifies the embedded default config keeps the
// "Opened with ..." info modal enabled, preserving prior behavior.
func TestShowOpenedModalDefault(t *testing.T) {
	v := viper.New()
	v.SetConfigType("toml")
	if err := v.ReadConfig(bytes.NewReader(defaultConf)); err != nil {
		t.Fatalf("failed to parse default config: %v", err)
	}
	if !v.GetBool("a-general.show_opened_modal") {
		t.Error("expected a-general.show_opened_modal to default to true")
	}
}

// TestShowOpenedModalOverride verifies the option can be disabled via config.
func TestShowOpenedModalOverride(t *testing.T) {
	v := viper.New()
	v.SetConfigType("toml")
	if err := v.ReadConfig(bytes.NewReader([]byte("[a-general]\nshow_opened_modal = false\n"))); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	if v.GetBool("a-general.show_opened_modal") {
		t.Error("expected a-general.show_opened_modal to be false when overridden")
	}
}
