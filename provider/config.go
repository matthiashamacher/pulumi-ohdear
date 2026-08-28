package provider

import (
	"context"
	"os"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// Config holds provider-level credentials for the Oh Dear API.
type Config struct {
	APIToken string `pulumi:"apiToken,optional" provider:"secret"`
}

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.APIToken, "Oh Dear API token. Falls back to the OHDEAR_API_TOKEN environment variable.")
}

// Configure fills in values from the environment when not set explicitly.
func (c *Config) Configure(ctx context.Context) error {
	if c.APIToken == "" {
		c.APIToken = os.Getenv("OHDEAR_API_TOKEN")
	}
	return nil
}
