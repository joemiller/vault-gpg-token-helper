APP := vault-gpg-token-helper

test:
	@go test -v ./...

build:
	@go build

build-linux:
	@GOOS=linux GOARCH=amd64 go build

release-snapshot:
	@rm -rf ./dist
	@goreleaser --snapshot

todo:
	@grep \
		--exclude-dir=vendor \
		--exclude-dir=dist \
		--text \
		--color \
		-nRo -E 'TODO:.*' .

.PHONY: build build-linux test release-snapshot todo
