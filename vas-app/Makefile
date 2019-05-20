SHELL:=/bin/bash
.PHONY:all install test lint

all: install lint test

lint:
	golangci-lint run

install:
	go install ./...

test:
	go test `go list ./... | grep -v qiniu.com/vas-app/app/dashboard | grep -v qiniu.com/vas-app/biz/dashboard/service`
