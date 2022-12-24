VERSION ?= $(shell git describe --tags --always --dirty)
GOFLAGS := -ldflags '-X main.version=$(VERSION)'
BIN     := 04-unusual-database

GOFILES := $(shell find . -name '*.go')

$(BIN): $(GOFILES) go.mod
	go build $(GOFLAGS) .

.PHONY: run
run:
	go run $(GOFLAGS) .

.PHONY: clean
clean:
	$(RM) $(BIN)
