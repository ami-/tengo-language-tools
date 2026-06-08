TENGOFMT_VERSION := 0.1.0
TENGOLS_VERSION  := 0.1.0
TENGO_VERSION    := $(shell go list -m -f '{{.Version}}' github.com/d5/tengo/v2)

TENGOFMT_LDFLAGS := -X main.version=$(TENGOFMT_VERSION) -X main.tengoVersion=$(TENGO_VERSION)
TENGOLS_LDFLAGS  := -X main.version=$(TENGOLS_VERSION)  -X main.tengoVersion=$(TENGO_VERSION)

.PHONY: all build clean

all: build

build:
	go build -ldflags "$(TENGOFMT_LDFLAGS)" -o tengofmt ./cmd/tengofmt
	go build -ldflags "$(TENGOLS_LDFLAGS)"  -o tengols  ./cmd/tengols

clean:
	rm -f tengofmt tengols
