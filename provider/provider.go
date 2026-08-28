package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
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
			infer.Resource(&Monitor{}),
			infer.Resource(&StatusPage{}),
			infer.Resource(&StatusPageUpdateTemplate{}),
			infer.Resource(&Tag{}),
			infer.Resource(&TagGroup{}),
		).
		WithConfig(infer.Config(&Config{})).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{"provider": "index"}).
		Build()
}
