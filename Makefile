.PHONY: build clean test package serve update-vendor
PKGS := $(shell go list ./... | grep -v /vendor/ |grep -v /api | grep -v /migrations | grep -v /static)
VERSION := $(shell git describe --always)
GOOS ?= linux
GOARCH ?= amd64

build:
	@echo "Compiling source for $(GOOS) $(GOARCH)"
	@mkdir -p build
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-X main.version=$(VERSION)" -o build/lora-app-server$(BINEXT) cmd/lora-app-server/main.go

clean:
	@echo "Cleaning up workspace"
	@rm -rf build
	@rm -rf dist/$(VERSION)

test:
	@echo "Running tests"
	@for pkg in $(PKGS) ; do \
		golint $$pkg ; \
	done
	@go vet $(PKGS)
	@go test -p 1 -v $(PKGS)

package: clean build
	@echo "Creating package for $(GOOS) $(GOARCH)"
	@mkdir -p dist/$(VERSION)
	@cp build/* dist/$(VERSION)
	@cd dist/$(VERSION) && tar -pczf ../lora_app_server_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz .
	@rm -rf dist/$(VERSION)

generate:
	@echo "Running go generate"

# shortcuts for development

serve: build
	@echo "Starting Lora App Server"
	./build/lora_app_server

update-vendor:
	@echo "Updating vendored packages"
	@govendor update +external

run-compose-test:
	docker-compose run --rm appserver make test
