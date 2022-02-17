OUT_DIR := out
PROG := docker-machine-driver-outscale

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

export GO111MODULE=on

ifeq ($(GOOS),windows)
	BIN_SUFFIX := ".exe"
endif

VERSION=$(shell git describe --exact-match 2> /dev/null || \
                 git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)

LDFLAGS=-ldflags "-X github.com/outscale-mdr/docker-machine-driver-outscale/pkg/drivers/outscale.version=${VERSION}"
.PHONY: build
build: dep
	go build $(LDFLAGS) -o $(OUT_DIR)/$(PROG)$(BIN_SUFFIX) ./

.PHONY: dep
dep:
	@GO111MODULE=on
	go get -d ./
	go mod verify

.PHONY: test
test: dep
	go test -race ./...

.PHONY: check
check:
	gofmt -l -s -d pkg/
	go vet

.PHONY: clean
clean:
	$(RM) $(OUT_DIR)/$(PROG)$(BIN_SUFFIX)

.PHONY: uninstall
uninstall:
	$(RM) $(GOPATH)/bin/$(PROG)$(BIN_SUFFIX)

.PHONY: install
install: dep
	go install $(LDFLAGS)

.PHONY: testacc
testacc: install
	@sh -c 'test/run_bats.sh test'