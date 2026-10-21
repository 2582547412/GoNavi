package provider

import (
	"GoNavi-Wails/internal/ai"
	"testing"
)

func pluginTestConfig() ai.ProviderConfig {
	return ai.ProviderConfig{ID: "selected", Type: "plugin:test.models:chat", Name: "Selected provider", Model: "model-1"}
}

func TestPluginConfigRejectsForeignAuthorityBeforeNormalization(t *testing.T) {
	if err := ValidatePluginConfig(pluginTestConfig()); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*ai.ProviderConfig){
		"namespace whitespace": func(c *ai.ProviderConfig) { c.Type = " plugin:test.models:chat" },
		"namespace case":       func(c *ai.ProviderConfig) { c.Type = "Plugin:test.models:chat" },
		"builtin id":           func(c *ai.ProviderConfig) { c.ID = " Gonavi-AI " },
		"api format":           func(c *ai.ProviderConfig) { c.APIFormat = "openai" },
		"cli path":             func(c *ai.ProviderConfig) { c.CLIPath = "x" },
		"cli env":              func(c *ai.ProviderConfig) { c.CLIEnv = map[string]string{"TOKEN": "secret"} },
		"effort":               func(c *ai.ProviderConfig) { c.Effort = "high" },
		"thinking":             func(c *ai.ProviderConfig) { c.ThinkingIntensity = "high" },
		"auth":                 func(c *ai.ProviderConfig) { c.AuthMode = "bearer" },
		"images":               func(c *ai.ProviderConfig) { v := true; c.SupportsImages = &v },
		"userinfo":             func(c *ai.ProviderConfig) { c.BaseURL = "https://secret@example.invalid" },
		"header duplicate":     func(c *ai.ProviderConfig) { c.Headers = map[string]string{"X-Key": "a", "x-key": "b"} },
		"header forbidden":     func(c *ai.ProviderConfig) { c.Headers = map[string]string{"Host": "example.invalid"} },
		"header newline":       func(c *ai.ProviderConfig) { c.Headers = map[string]string{"X-Key": "a\nb"} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c := pluginTestConfig()
			mutate(&c)
			if err := ValidatePluginConfig(c); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
}
