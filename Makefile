.PHONY: build clean test ui-requirements serve internal/statics internal/migrations 
PKGS := $(shell go list ./... | grep -v /vendor |grep -v chirpstack-application-server/api | grep -v /migrations | grep -v /static | grep -v /ui)
VERSION := $(shell git describe --always |sed -e "s/^v//")
API_VERSION := $(shell go list -m -f '{{ .Version }}' github.com/brocaar/chirpstack-api/go/v3 | awk '{n=split($$0, a, "-"); print a[n]}')

build: ui/build internal/statics internal/migrations
	mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/chirpstack-application-server cmd/chirpstack-application-server/main.go

clean:
	@echo "Cleaning up workspace"
	@rm -rf build dist internal/migrations/migrations_gen.go internal/static/static_gen.go ui/build static/static
	@rm -f static/index.html static/icon.png static/manifest.json static/asset-manifest.json static/service-worker.js
	@rm -rf static/logo
	@rm -rf static/swagger/*.json
	@rm -rf docs/public
	@rm -rf dist

test: internal/statics internal/migrations
	@echo "Running tests"
	@rm -f coverage.out
	@for pkg in $(PKGS) ; do \
		golint $$pkg ; \
	done
	@go vet $(PKGS)
	@go test -p 1 -v $(PKGS) -cover -coverprofile coverage.out

dist: ui/build internal/statics internal/migrations
	@goreleaser
	mkdir -p dist/upload/tar
	mkdir -p dist/upload/deb
	mkdir -p dist/upload/rpm
	mv dist/*.tar.gz dist/upload/tar
	mv dist/*.deb dist/upload/deb
	mv dist/*.rpm dist/upload/rpm

snapshot: ui/build internal/statics internal/migrations
	@goreleaser --snapshot

ui/build:
	@echo "Building ui"
	@cd ui && npm run build
	@mv ui/build/* static

internal/statics internal/migrations: static/swagger/api.swagger.json
	@echo "Generating static files"
	@go generate internal/migrations/migrations.go
	@go generate internal/static/static.go


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
	go install github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs
	go install github.com/jteeuwen/go-bindata/go-bindata
	go install golang.org/x/tools/cmd/stringer
	go install github.com/goreleaser/goreleaser
	go install github.com/goreleaser/nfpm

ui-requirements:
	@echo "Installing UI requirements"
	@cd ui && npm install

serve: build
	@echo "Starting ChirpStack Application Server"
	./build/chirpstack-application-server
