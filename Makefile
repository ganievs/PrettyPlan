.PHONY: build
build:
	go build -o dist/pp cmd/prettyplan/main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...
