GO = go

default: run

run: build
	./bin/wakeonlan

build: embedFS/ go.sum main.go
	$(GO) build -o bin/

buildpi: embedFS/ go.sum main.go
	$Env:GOOS="linux" ; $Env:GOARCH="arm64" ; $(GO) build -o bin/

deps: go.sum embedfs/

go.sum: go.mod
	$(GO) mod tidy
	$(GO) mod download
	$(GO) mod verify
	$(GO) get github.com/UnnoTed/fileb0x


embedFS/: views/ assets/ go.sum
	$(GO) generate

