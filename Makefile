PROJECT_ROOT := $(shell pwd)

.PHONY: vet
vet:
	@echo "Vet..."
	@go vet `go list $(PROJECT_ROOT)/...`

.PHONY: build
build: vet
	go build -p 4 -o build/bin/gobore cmd/main.go

.PHONY: server
server: build
	./build/bin/gobore server --secret 123

.PHONY: client
client: build
	./build/bin/gobore client --local-host 127.0.0.1 --local-port 22 --to 127.0.0.1 --secret 123