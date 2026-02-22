# Makefile for common dev flows
.PHONY: build run vet lint fmt clean

build:
	go build -o recliner main.go

run: build
	./recliner

vet:
	go vet ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

clean:
	rm -f recliner
