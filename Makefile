.PHONY: test race cover fmt fmt-check vet example tidy check

test:
	@go test -shuffle=on ./...

race:
	@go test -race -shuffle=on -count=1 ./...

cover:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1
	@go tool cover -html=coverage.out -o coverage.html

fmt:
	@gofmt -l -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || { echo "unformatted:"; gofmt -l .; exit 1; }

vet:
	@go vet ./...

example:
	@cd _example && go build ./...

tidy:
	@go mod tidy
	@cd _example && go mod tidy

check: fmt-check vet race example
