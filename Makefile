## Makefile

.DEFAULT_GOAL := build
BUILD_NUMBER ?= SNAPSHOT-$(shell git rev-parse --abbrev-ref HEAD)

GOOS ?= $(uname -s)
GOARCH ?= amd64
export GOOS
export GOARCH

## Build a statically linked binary using a Docker container
BUILD_APP_PATH = /gopath/src/github.com/starlingbank/$(shell basename $(shell pwd))

build: clean get
	docker run --rm -t -v "$(GOPATH)":/gopath -v "$(shell pwd)":"$(BUILD_APP_PATH)" -e "GOPATH=/gopath" -w $(BUILD_APP_PATH) golang:1.9.2-alpine3.7 sh -c 'CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s"'

clean:
	go clean

install: clean
	go get -t .
	go install .

test: get-deps-tests get
	go test ./...

get:
	go get -t .

get-deps-tests:
	@echo "go get testing dependencies"
	go get github.com/stretchr/testify

docker:
	docker build -t quay.io/starlingbank/vaultsmith:$(BUILD_NUMBER) .
	docker build -t quay.io/starlingbank/vaultsmith:latest .

ifneq ($(findstring SNAPSHOT, $(BUILD_NUMBER)), SNAPSHOT)
  ifeq ($(shell git rev-parse --abbrev-ref HEAD), master)
publish: docker
	docker push quay.io/starlingbank/vaultsmith:latest
	docker push quay.io/starlingbank/vaultsmith:$(BUILD_NUMBER)
  else
publish:
	$(warning skipping target "publish", not on master branch)
  endif
else
publish:
	$(error the target "publish" requires that BUILD_NUMBER be set)
endif

.PHONY: install, build, docker, test, publish
