deps:
	go mod download

go_build: deps
	GOOS=linux GOARCH=amd64 go build

test:
	go test ./...

build: go_build test
