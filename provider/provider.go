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
		WithNamespace("matthiashamacher").
		WithDisplayName("Oh Dear").
		WithDescription("A Pulumi provider for Oh Dear: manage monitors, status pages, tags, and notification destinations.").
		WithKeywords("ohdear", "monitoring", "uptime", "status-page", "category/monitoring", "kind/native").
		WithHomepage("https://ohdear.app").
		WithRepository("https://github.com/matthiashamacher/pulumi-ohdear").
		WithGoImportPath("github.com/matthiashamacher/pulumi-ohdear/sdk/go/ohdear").
		WithPublisher("Matthias Hamacher").
		WithLicense("Apache-2.0").
		WithLogoURL("https://raw.githubusercontent.com/matthiashamacher/pulumi-ohdear/main/assets/logo.png").
		WithPluginDownloadURL("github://api.github.com/matthiashamacher/pulumi-ohdear").
		WithResources(
			infer.Resource(&Site{}),
			infer.Resource(&Monitor{}),
			infer.Resource(&CronCheck{}),
			infer.Resource(&RecurringMaintenancePeriod{}),
			infer.Resource(&NotificationDestination{}),
			infer.Resource(&StatusPage{}),
			infer.Resource(&Tag{}),
			infer.Resource(&TagGroup{}),
		).
		WithConfig(infer.Config(&Config{})).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{"provider": "index"}).
		// WithLanguageMap replaces the builder's defaults wholesale, so the
		// go/csharp entries below are copied verbatim from
		// infer.NewProviderBuilder. The additions:
		//  - nodejs/python packageName: `gen-sdk` derives these from the
		//    namespace, but the bare schema omits them and the registry's
		//    publish probe needs the names we actually push.
		//  - java basePackage: without it `gen-sdk --language java` derives a
		//    malformed `com.matthiashamacher.` Maven group from the namespace.
		WithLanguageMap(map[string]any{
			"nodejs": map[string]any{
				"respectSchemaVersion": true,
				"packageName":          "@matthiashamacher/ohdear",
			},
			"go": map[string]any{
				"generateResourceContainerTypes": true,
				"respectSchemaVersion":           true,
			},
			"python": map[string]any{
				"respectSchemaVersion": true,
				"packageName":          "matthiashamacher_ohdear",
				"pyproject": map[string]any{
					"enabled": true,
				},
			},
			"csharp": map[string]any{
				"respectSchemaVersion": true,
			},
			"java": map[string]any{
				"basePackage": "io.github.matthiashamacher",
			},
		}).
		Build()
}
