#!/usr/bin/make -f
tidy:
	go mod tidy

fmt:
	go fmt ./...

update:
	go get -u ./...
	go mod tidy

target:
	go build ./...

test:
	go test ./...
