PROJECT_ROOT=$(shell pwd)
VENDOR_PATH=$(PROJECT_ROOT)/vendor

GOPATH=$(PROJECT_ROOT):$(VENDOR_PATH)
export GOPATH

all: 
	@go build -o test/route53 test/route53.go

fmt:
	@find src -name \*.go -exec gofmt -l -w {} \;
	@find test -name \*.go -exec gofmt -l -w {} \;

clean:
	@rm -f test/route53
