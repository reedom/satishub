GO		= go
GOX		= gox
RM		= rm

VERSION  := $(shell git describe --tags)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := \
	-X 'github.com/reedom/satishub/cmd.version=$(VERSION)' \
	-X 'github.com/reedom/satishub/cmd.revision=$(REVISION)'
LDFLAGS_DIST := -s -w $(LDFLAGS)

TARGET	= satishub
SRC     := $(shell find . -name '*.go' | egrep -v "\/_?vendor")
PKGS	:= . ./api/... ./cmd/... ./pkg/...

# .SILENT: ;               # no need for @
.ONESHELL: ;             # recipes execute in same shell
.NOTPARALLEL: ;          # wait for this target to finish
.EXPORT_ALL_VARIABLES: ; # send all vars to shell

default: help-default    # default target
Makefile: ;              # skip prerequisite discovery

.title:
	$(info satis-hub manager $(VERSION))
	$(info )

help-default help: .title
	@echo "             setup: install required tools"
	@echo "              deps: install dependant packages"
	@echo "               fmt: re-format go source code files"
	@echo "              lint: run golint"
	@echo "              test: run tests"
	@echo "             build: Build satishub for development on the current environment"
	@echo "              dist: Build satishub binaries for distribution"
	@echo "      docker-build: create a new docker image"
	@echo ""

setup:
	go get github.com/golang/dep/cmd/dep
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports
#	go get github.com/Songmu/make2help/cmd/make2help

deps: setup
	dep ensure -v

lint:
	go vet $(PKGS)
	$(foreach f, $(SRC), golint $(f);)

fmt:
	goimports -w $(SRC)

test:
	go test $(PKGS)

build: bin/$(TARGET)

bin/$(TARGET): $(SRC)
	$(GO) build -o bin/$(TARGET) -ldflags "$(LDFLAGS)" main.go

dist: bin/linux_amd64/$(TARGET) bin/darwin_amd64/$(TARGET)

bin/linux_amd64/$(TARGET): $(SRC)
	$(GOX) \
		-osarch="linux/amd64" \
		-ldflags "$(LDFLAGS_DIST)" \
		-output="bin/linux_amd64/$(TARGET)"

bin/darwin_amd64/$(TARGET): $(SRC)
	$(GOX) \
		-osarch="darwin/amd64" \
		-ldflags "$(LDFLAGS_DIST)" \
		-output="bin/darwin_amd64/$(TARGET)"

docker-build:
	docker build --tag reedom/satishub:$(VERSION) --build-arg LDFLAGS="$(LDFLAGS_DIST)" .
	docker tag reedom/satishub:$(VERSION) reedom/satishub:latest
