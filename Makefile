deps:
	@go get

lint:
	@golangci-lint run -v

test:
	@go test -v ./...

build:
	@CGO_ENABLED=0 go build .

build-linux:
	@CGO_ENABLED=0 GOOS=linux go build .

snapshot:
	@goreleaser --snapshot --rm-dist --debug

todo:
	@grep \
		--exclude-dir=vendor \
		--exclude-dir=dist \
		--text \
		--color \
		-nRo -E 'TODO:.*' .

.PHONY: build build-linux test snapshot todo