.PHONY: update-dependencies
update-dependencies:
	@go get -t -u ./... && go mod tidy

.PHONY: format
format:
	goimports -local github.com/justcodes/loafer-go -w -l .

.PHONY: lint
lint:
	@$(MAKE) format
	@golangci-lint run --allow-parallel-runners ./... --max-same-issues 0

.PHONY: install-golang-ci
install-golang-ci:
	@echo "Installing golang-ci"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1
	@echo "Golang-ci installed successfully"

.PHONY: install-goimports
install-goimports:
	@echo "Installing go imports"
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Go imports installed successfully"

.PHONY: configure
configure:
	make install-golang-ci
	make install-goimports

.PHONY: clean
clean:
	@go clean -testcache

.PHONY: test
test:
	@$(MAKE) -s clean
	@go test -timeout 1m -race -covermode=atomic -coverprofile=tmp.out ./... && cat tmp.out | grep -Ev 'example|fake' > geral.out  && go tool cover -func=geral.out
