PROVIDER := pulumi-resource-ohdear
VERSION  := $(shell (git describe --tags --match 'v*' --dirty 2>/dev/null || echo v0.1.0) | sed 's/^v//')
LDFLAGS  := -ldflags "-X github.com/matthiashamacher/pulumi-ohdear/provider.Version=$(VERSION)"

.PHONY: build install schema sdk sdk_all dist clean

build:
	go build $(LDFLAGS) -o bin/$(PROVIDER) ./provider/cmd/$(PROVIDER)

install: build
	cp bin/$(PROVIDER) $(shell go env GOPATH)/bin/

schema: build
	pulumi package get-schema ./bin/$(PROVIDER) > schema.json

sdk: build
	rm -rf sdk
	pulumi package gen-sdk ./bin/$(PROVIDER) --language nodejs -o sdk

# All languages, for a full Registry listing.
sdk_all: build
	rm -rf sdk
	pulumi package gen-sdk ./bin/$(PROVIDER) --language nodejs,python,dotnet,go,java -o sdk

# Local snapshot build of the plugin archives (no publish).
dist:
	goreleaser release --snapshot --clean

clean:
	rm -rf bin sdk schema.json dist
