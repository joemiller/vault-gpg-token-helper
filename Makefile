APP := vault-gpg-token-helper

test:
	@go test -v ./...

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

.PHONY: test release-snapshot todo
