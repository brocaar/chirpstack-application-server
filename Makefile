.PHONY: build clean test package serve update-vendor ui
PKGS := $(shell go list ./... | grep -v /vendor/ |grep -v /api | grep -v /migrations | grep -v /static | grep -v /ui)
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

ui:
	@echo "Building ui"
	@rm -f static/index.html
	@rm -rf static/static
	@rm -rf ui/build
	@cd ui && npm run build
	@mv ui/build/* static
	@echo "Don't forget to run make generate to include the static files in the Go code!"

generate:
	@echo "Running go generate"
	@go generate api/api.go
	@go generate cmd/lora-app-server/main.go

# shortcuts for development

requirements:
	@echo "Installing development tools"
	@go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	@go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@go get -u github.com/elazarl/go-bindata-assetfs/...

serve: build
	@echo "Starting Lora App Server"
	./build/lora-app-server

update-vendor:
	@echo "Updating vendored packages"
	@govendor update +external

run-compose-test:
	docker-compose run --rm appserver make test
