.PHONY: build install test clean release deps

VERSION := 1.0.0
CODENAME := Anchor
BINARY := kselect
LDFLAGS := -ldflags="-X main.Version=$(VERSION) -X main.Codename=$(CODENAME)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/kselect/

install: build
	sudo mv $(BINARY) /usr/local/bin/

test:
	go test -v ./...

clean:
	rm -f $(BINARY) $(BINARY)-*
	go clean

deps:
	go mod tidy

release:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 ./cmd/kselect/
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 ./cmd/kselect/
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 ./cmd/kselect/
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-darwin-arm64 ./cmd/kselect/
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe ./cmd/kselect/
