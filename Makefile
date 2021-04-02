.PHONY: build clean test ui-requirements serve statics
VERSION := $(shell git describe --always |sed -e "s/^v//")
API_VERSION := $(shell go list -m -f '{{ .Version }}' github.com/brocaar/chirpstack-api/go/v3 | awk '{n=split($$0, a, "-"); print a[n]}')

build: ui/build static/swagger/api.swagger.json
	mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/chirpstack-application-server cmd/chirpstack-application-server/main.go

clean:
	@echo "Cleaning up workspace"
	@rm -rf build dist internal/static/static_gen.go ui/build static/static
	@rm -f static/index.html static/icon.png static/manifest.json static/asset-manifest.json static/service-worker.js
	@rm -rf static/logo
	@rm -rf static/integrations
	@rm -rf static/swagger/*.json
	@rm -rf dist

test: statics
	@echo "Running tests"
	@rm -f coverage.out
	@golint ./...
	@go vet ./...
	@go test -p 1 -v -cover ./... -coverprofile coverage.out

dist: statics
	@goreleaser
	mkdir -p dist/upload/tar
	mkdir -p dist/upload/deb
	mkdir -p dist/upload/rpm
	mv dist/*.tar.gz dist/upload/tar
	mv dist/*.deb dist/upload/deb
	mv dist/*.rpm dist/upload/rpm

snapshot: statics
	@goreleaser --snapshot

proto:
	@rm -rf /tmp/chirpstack-api
	@git clone https://github.com/brocaar/chirpstack-api.git /tmp/chirpstack-api
	@git --git-dir=/tmp/chirpstack-api/.git --work-tree=/tmp/chirpstack-api checkout $(API_VERSION)
	@go generate internal/integration/loracloud/frame_rx_info.go

statics: ui/build static/swagger/api.swagger.json

ui/build:
	@echo "Building ui"
	@cd ui && npm run build
	@mv ui/build/* static

static/swagger/api.swagger.json:
	@echo "Fetching Swagger definitions and generate combined Swagger JSON"
	@rm -rf /tmp/chirpstack-api
	@git clone https://github.com/brocaar/chirpstack-api.git /tmp/chirpstack-api
	@git --git-dir=/tmp/chirpstack-api/.git --work-tree=/tmp/chirpstack-api checkout $(API_VERSION)
	@mkdir -p static/swagger
	@cp /tmp/chirpstack-api/swagger/as/external/api/*.json static/swagger
	@GOOS="" GOARCH="" go run internal/tools/swagger/main.go /tmp/chirpstack-api/swagger/as/external/api > static/swagger/api.swagger.json

# shortcuts for development

dev-requirements:
	go mod download
	go install golang.org/x/lint/golint
	go install golang.org/x/tools/cmd/stringer
	go install github.com/goreleaser/goreleaser
	go install github.com/goreleaser/nfpm
	go install github.com/golang/protobuf/protoc-gen-go
	go install github.com/golang-migrate/migrate/v4/cmd/migrate

ui-requirements:
	@echo "Installing UI requirements"
	@cd ui && npm install

serve: build
	@echo "Starting ChirpStack Application Server"
	./build/chirpstack-application-server
