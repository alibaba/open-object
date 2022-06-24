# go parameters
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test -v
GO_PACKAGE=github.com/alibaba/open-object
GO_ARCH=amd64
# GO_OS=darwin
GO_OS=linux

# build info
NAME=open-object
OUTPUT_DIR=./bin
MAIN_FILE=./cmd/main.go

IMAGE_NAME=thebeatles1994/${NAME}
VERSION=v0.1.0-dev
IMAGE_TAG=${IMAGE_NAME}:${VERSION}
GIT_COMMIT=$(shell git rev-parse HEAD)
LD_FLAGS=-ldflags "-X '${GO_PACKAGE}/pkg/version.GitCommit=$(GIT_COMMIT)' -X '${GO_PACKAGE}/pkg/version.Version=$(VERSION)' -X 'main.VERSION=$(VERSION)' -X 'main.COMMITID=$(GIT_COMMIT)'"

.PHONY: test build container sync clean

.PHONY: build
build:
	GOARCH=$(GO_ARCH) GOOS=$(GO_OS) CGO_ENABLED=0 $(GO_BUILD) $(LD_FLAGS) -v -o $(OUTPUT_DIR)/$(NAME) $(MAIN_FILE)
	chmod +x $(OUTPUT_DIR)/$(NAME)

local:
	GO111MODULE=off CGO_ENABLED=0 go build $(LD_FLAGS) -v -o $(BIN_DRIVER_NAME) . .
image: build
	chmod +x $(OUTPUT_DIR)/$(NAME)
	chmod +x build/run-connector.sh
	docker build -t $(IMAGE_TAG) -f build/Dockerfile .
	docker push $(IMAGE_TAG)
clean:
	go clean -r -x
	-rm -rf _output
