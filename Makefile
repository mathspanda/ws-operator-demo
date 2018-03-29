GODIR = $(shell go list ./... | grep -v /vendor/)
PKG := github.com/mathspanda/ws-operator-demo
GOARCH := amd64
GOOS := linux
BUILD_IMAGE ?= golang:1.9.0-alpine

build-dirs:
	@mkdir -p .go/src/$(PKG) ./go/bin
	@mkdir -p release
.PHONY: build-dirs

build-operator: build-dirs
	@docker run                                                            \
	    --rm                                                               \
	    -ti                                                                \
	    -u $$(id -u):$$(id -g)                                             \
	    -v $$(pwd)/.go:/go                                                 \
	    -v $$(pwd):/go/src/$(PKG)                                          \
	    -v $$(pwd)/release:/go/bin                                         \
	    -e GOOS=$(GOOS)                                                    \
	    -e GOARCH=$(GOARCH)                                                \
	    -e CGO_ENABLED=0                                                   \
	    -w /go/src/$(PKG)                                                  \
	    $(BUILD_IMAGE)                                                     \
	    go install -v -pkgdir /go/pkg ./cmd/operator
.PHONY: build-operator

operator-image: build-operator
	@sh build/build_operator.sh
.PHONY: operator-image

ws-image:
	@docker run                                                            \
		--rm                                                               \
	    -ti                                                                \
		-u $$(id -u):$$(id -g)                                             \
		-v $$(pwd)/dockerfile/webserver:/tmp                               \
		-e GOOS=$(GOOS)                                                    \
		-e GOARCH=$(GOARCH)                                                \
		-e CGO_ENABLED=0                                                   \
		-w /tmp                                                            \
		$(BUILD_IMAGE)                                                     \
		go build simple_server.go
	@sh build/build_simple_ws.sh
.PHONY: ws-image
