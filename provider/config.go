package provider

import (
	"context"
	"errors"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// Config holds provider-level credentials for the Oh Dear API.
type Config struct {
	APIToken string `pulumi:"apiToken,optional" provider:"secret"`

	client *Client
}

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.APIToken, "Oh Dear API token, sent as a Bearer token on every request. "+
		"Defaults to the OHDEAR_API_TOKEN environment variable.")
	a.SetDefault(&c.APIToken, "", "OHDEAR_API_TOKEN")
}

// Configure validates the token and builds the shared API client.
func (c *Config) Configure(ctx context.Context) error {
	if c.APIToken == "" {
		return errors.New("no Oh Dear API token: set the `apiToken` provider config or the OHDEAR_API_TOKEN environment variable")
	}
	c.client = NewClient(c.APIToken)
	return nil
}
