PROJECT_ROOT=$(shell pwd)
VENDOR_PATH=$(PROJECT_ROOT)/vendor

GOPATH=$(PROJECT_ROOT):$(VENDOR_PATH)
export GOPATH

all: cli

cli:
	@go build -o test/cli test/cli.go

fmt:
	@find src -name \*.go -exec gofmt -l -w {} \;

clean:
	@rm -f test/cli
