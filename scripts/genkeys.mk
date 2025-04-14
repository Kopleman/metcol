BUILD_DATE = $(shell date +'%Y/%m/%d %H:%M:%S')
BUILD_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_VERSION = $(shell git describe --always --long --dirty)

.PHONY: build-generator
build-generator:
	go build  -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit=$(BUILD_COMMIT)'" -o ./cmd/generatekeys/generatekeys ./cmd/generatekeys

.PHONY: generatekeys
generatekeys:
	./cmd/generatekeys/generatekeys -p "public.pem" -r "private.pem"