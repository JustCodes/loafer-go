.PHONY: update-dependencies
update-dependencies:
	@go get -t -u ./... && go mod tidy
