go_build:
	GOOS=linux GOARCH=amd64 go build

test:
	go test ./...

build: go_build test
