default: version test build

fmt:
	go fmt github.com/DimensionDataResearch/dd-tf-export

# Peform a development (current-platform-only) build.
dev: fmt
	go build -o _bin/dd-tf-export

install: dev
	go install

# Perform a full (all-platforms) build.
build: version build-windows64 build-linux64 build-mac64

build-windows64:
	GOOS=windows GOARCH=amd64 go build -o _bin/windows-amd64/dd-tf-export.exe

build-linux64:
	GOOS=linux GOARCH=amd64 go build -o _bin/linux-amd64/dd-tf-export

build-mac64:
	GOOS=darwin GOARCH=amd64 go build -o _bin/darwin-amd64/dd-tf-export

# Produce archives for a GitHub release.
dist: build
	zip -9 _bin/windows-amd64.zip _bin/windows-amd64/dd-tf-export.exe
	zip -9 _bin/linux-amd64.zip _bin/linux-amd64/dd-tf-export
	zip -9 _bin/darwin-amd64.zip _bin/darwin-amd64/dd-tf-export

test: fmt
	go test -v github.com/DimensionDataResearch/dd-tf-export

version:
	echo "package main\n\n// ProgramVersion is the current version of the DD Cloud Compute terraform exporter.\nconst ProgramVersion = \"v0.1 (`git rev-parse HEAD`)\"" > ./version-info.go
