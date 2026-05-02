GO := go
WIRE := wire
SWAG := swag
BINARY_NAME := api_gateway

.PHONY: build test wire swagger clean

build: wire swagger
	$(GO) build -o $(BINARY_NAME) main.go

test:
	$(GO) test ./...

wire:
	$(WIRE) ./src/infra/di

swagger:
	$(SWAG) init -g main.go --parseDependency --parseInternal

clean:
	rm -f $(BINARY_NAME)
