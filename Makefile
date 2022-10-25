PROJECT_ROOT := $(shell pwd)
VERSION :=1.1.5

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
	./build/bin/gobore client --local-host 10.11.0.213 --local-port 22 --to 127.0.0.1 --secret 123

image:
	GOOS=linux GOARCH=amd64 go build -p 4 -o build/bin/gobore cmd/main.go
	docker build -f build/Dockerfile -t mqyang56/gobore:$(VERSION) .