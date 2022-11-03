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
VERSION=v0.1.0
IMAGE_NAME_FOR_DOCKERHUB=thebeatles1994/${NAME}
GIT_COMMIT=$(shell git rev-parse HEAD)
LD_FLAGS=-ldflags "-X '${GO_PACKAGE}/pkg/version.GitCommit=$(GIT_COMMIT)' -X '${GO_PACKAGE}/pkg/version.Version=$(VERSION)' -X 'main.VERSION=$(VERSION)' -X 'main.COMMITID=$(GIT_COMMIT)'"

.PHONY: build image clean

.PHONY: build
build:
	CGO_ENABLED=0 $(GO_BUILD) $(LD_FLAGS) -v -o $(OUTPUT_DIR)/$(NAME) $(MAIN_FILE)

.PHONY: develop
develop:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 $(GO_BUILD) $(LD_FLAGS) -v -o $(OUTPUT_DIR)/$(NAME) $(MAIN_FILE)
	chmod +x $(OUTPUT_DIR)/$(NAME)
	docker build . -t ${IMAGE_NAME_FOR_DOCKERHUB}:${VERSION} -f ./build/Dockerfile.dev

# build image
.PHONY: image
image:
	docker build . -t ${IMAGE_NAME_FOR_DOCKERHUB}:${VERSION} -f ./build/Dockerfile

# build image for arm64
.PHONY: image-arm64
image-arm64:
	docker build . -t ${IMAGE_NAME_FOR_DOCKERHUB}:${VERSION}-arm64 -f ./build/Dockerfile.arm64

clean:
	go clean -r -x
	-rm -rf bin
