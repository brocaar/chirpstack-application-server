.PHONY: build clean test package package-deb ui api statics requirements ui-requirements serve update-vendor internal/statics internal/migrations static/swagger/api.swagger.json
PKGS := $(shell go list ./... | grep -v /vendor |grep -v lora-app-server/api | grep -v /migrations | grep -v /static | grep -v /ui)
VERSION := $(shell git describe --always)
GOOS ?= linux
GOARCH ?= amd64

build: ui/build internal/statics internal/migrations
	@echo "Compiling source for $(GOOS) $(GOARCH)"
	@mkdir -p build
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/lora-app-server$(BINEXT) cmd/lora-app-server/main.go

clean:
	@echo "Cleaning up workspace"
	@rm -rf build dist internal/migrations internal/static ui/build static/static
	@rm -f static/index.html
	@rm -rf docs/public

test: internal/statics internal/migrations
	@echo "Running tests"
	@for pkg in $(PKGS) ; do \
		golint $$pkg ; \
	done
	@go vet $(PKGS)
	@go test -p 1 -v $(PKGS)

documentation:
	@echo "Building documentation"
	@mkdir -p dist/docs
	@cd docs && hugo
	@cd docs/public/ && tar -pczf ../../dist/docs/lora-app-server.tar.gz .

package: build
	@echo "Creating package for $(GOOS) $(GOARCH)"
	@mkdir -p dist/tar/$(VERSION)
	@cp build/* dist/tar/$(VERSION)
	@cd dist/tar/$(VERSION) && tar -pczf ../lora_app_server_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz .
	@rm -rf dist/tar/$(VERSION)

package-deb: package
	@echo "Building deb package for $(GOOS) $(GOARCH)"
	@cd packaging && TARGET=deb ./package.sh

ui/build:
	@echo "Building ui"
	@cd ui && npm run build
	@mv ui/build/* static

api:
	@echo "Generating API code from .proto files"
	@go generate api/api.go

internal/statics internal/migrations: static/swagger/api.swagger.json
	@echo "Generating static files"
	@go generate cmd/lora-app-server/main.go


static/swagger/api.swagger.json:
	@echo "Generating combined Swagger JSON"
	@GOOS="" GOARCH="" go run api/swagger/main.go api/swagger > static/swagger/api.swagger.json


# shortcuts for development

requirements:
	@echo "Installing development tools"
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	@go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@go get -u github.com/elazarl/go-bindata-assetfs/...
	@go get -u github.com/jteeuwen/go-bindata/...
	@go get -u github.com/kisielk/errcheck
	@go get -u github.com/smartystreets/goconvey
	@go get -u golang.org/x/tools/cmd/stringer
	@go get -u github.com/golang/dep/cmd/dep
	@dep ensure -v

ui-requirements:
	@echo "Installing UI requirements"
	@cd ui && npm install

serve: build
	@echo "Starting Lora App Server"
	./build/lora-app-server

update-vendor:
	@echo "Updating vendored packages"
	@govendor update +external

run-compose-test:
	docker-compose run --rm appserver make test
