package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// Version is set via -ldflags at build time.
var Version = "0.1.0"

const Name = "ohdear"

// New builds the Oh Dear Pulumi provider.
func New() (p.Provider, error) {
	return infer.NewProviderBuilder().
		WithNamespace("mhamacher").
		WithResources(
			infer.Resource(&Site{}),
		).
		WithConfig(infer.Config(&Config{})).
		Build()
}
