COMMANDS = addict hangover makeodbc n2j odbcd pbfmt
DIST = $(COMMANDS:%=dist/%)
.PHONY = $(DIST) all dist deps clean test

LINUX_AMD64 = $(COMMANDS:%=dist/%_linux_amd64)
DARWIN = $(COMMANDS:%=dist/%_darwin)
DARWIN_AMD64 = $(COMMANDS:%=dist/%_darwin_amd64)
DARWIN_ARM64 = $(COMMANDS:%=dist/%_darwin_arm64)
WINDOWS_AMD64 = $(COMMANDS:%=dist/%_windows_amd64.exe)

all: $(DIST)
$(DIST): dist/%: dist/%_linux_amd64 dist/%_darwin dist/%_windows_amd64.exe

$(LINUX_AMD64):dist/%_linux_amd64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ ./cmd/$*/
$(DARWIN):dist/%_darwin: dist/%_darwin_arm64 dist/%_darwin_amd64
	mkdir -p dist
	lipo -output $@ -create dist/$*_darwin_arm64 dist/$*_darwin_amd64
$(DARWIN_ARM64):dist/%_darwin_arm64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $@ ./cmd/$*/
$(DARWIN_AMD64):dist/%_darwin_amd64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ ./cmd/$*/
$(WINDOWS_AMD64):dist/%_windows_amd64.exe:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $@ ./cmd/$*/

clean:
	rm -rf dist

generate:
	go generate ./...

test:
	go test ./...

deps: internal/assets/node_modules
	go get ./...
	go mod tidy
	go install github.com/tkw1536/lipo@latest
	go install github.com/tkw1536/gogenlicense/cmd/gogenlicense@latest

internal/assets/node_modules:
	cd internal/assets/ && yarn install