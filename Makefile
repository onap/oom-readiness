go_build:
	GOOS=linux GOARCH=amd64 go build

test:
	go test ./...

build: test go_build
