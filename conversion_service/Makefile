GO := go
WIRE := wire
BINARY_NAME := conversion_worker

.PHONY: build test wire clean

build: wire
	$(GO) build -o $(BINARY_NAME) main.go

test:
	$(GO) test ./...

wire:
	$(WIRE) ./src/infra/di

clean:
	rm -f $(BINARY_NAME)
