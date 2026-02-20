package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration, read from environment variables.
// All fields are unexported — use accessor methods for read-only access.
type Config struct {
	port             string
	cloverBaseURL    string
	cloverEcomURL    string
	cloverMerchant   string
	cloverPublicKey  string
	cloverPrivateKey string
}

// Load reads environment variables, validates required ones, and returns a
// readonly Config. It returns an error listing any missing required variables.
func Load() (*Config, error) {
	required := map[string]*string{
		"CLOVER_BASE_URL":          new(string),
		"CLOVER_ECOMMERCE_URL":     new(string),
		"CLOVER_MERCHANT_ID":       new(string),
		"CLOVER_PUBLIC_API_TOKEN":  new(string),
		"CLOVER_PRIVATE_API_TOKEN": new(string),
	}

	var missing []string
	for name, ptr := range required {
		val := os.Getenv(name)
		if val == "" {
			missing = append(missing, name)
		}
		*ptr = val
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		port:             port,
		cloverBaseURL:    *required["CLOVER_BASE_URL"],
		cloverEcomURL:    *required["CLOVER_ECOMMERCE_URL"],
		cloverMerchant:   *required["CLOVER_MERCHANT_ID"],
		cloverPublicKey:  *required["CLOVER_PUBLIC_API_TOKEN"],
		cloverPrivateKey: *required["CLOVER_PRIVATE_API_TOKEN"],
	}, nil
}

func (c *Config) Port() string               { return c.port }
func (c *Config) CloverBaseURL() string      { return c.cloverBaseURL }
func (c *Config) CloverEcommerceURL() string { return c.cloverEcomURL }
func (c *Config) CloverMerchantID() string   { return c.cloverMerchant }
func (c *Config) CloverPublicKey() string    { return c.cloverPublicKey }
func (c *Config) CloverPrivateKey() string   { return c.cloverPrivateKey }

// CloverSDKURL returns the Clover hosted checkout SDK URL, derived from the base URL.
// Sandbox base URLs produce the sandbox SDK; all others produce the production SDK.
func (c *Config) CloverSDKURL() string {
	if strings.Contains(c.cloverBaseURL, "sandbox.dev.clover.com") {
		return "https://checkout.sandbox.dev.clover.com/sdk.js"
	}
	return "https://checkout.clover.com/sdk.js"
}
