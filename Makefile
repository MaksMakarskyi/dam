.PHONY: test race cover fmt fmt-check vet example damecho damgin tidy check \
	version-guard publish publish-adapters

VERSION  ?=
MODULE   := github.com/MaksMakarskyi/dam
ADAPTERS := damecho damgin

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

damecho:
	@cd damecho && go build ./... && go vet ./... && go test ./...

damgin:
	@cd damgin && go build ./... && go vet ./... && go test ./...

tidy:
	@go mod tidy
	@cd _example && go mod tidy
	@cd damecho && go mod tidy
	@cd damgin && go mod tidy

check: fmt-check vet race example damecho damgin

version-guard:
	@test -n "$(VERSION)" || { echo "usage: make $(MAKECMDGOALS) VERSION=vX.Y.Z"; exit 1; }
	@echo "$(VERSION)" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$$' \
		|| { echo "VERSION must look like v1.2.3, got: $(VERSION)"; exit 1; }
	@git diff --quiet HEAD || { echo "working tree is dirty; commit first"; exit 1; }

publish: version-guard check
	@! git rev-parse -q --verify refs/tags/$(VERSION) >/dev/null \
		|| { echo "tag $(VERSION) already exists"; exit 1; }
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

publish-adapters: version-guard
	@git ls-remote --exit-code --tags origin refs/tags/$(VERSION) >/dev/null \
		|| { echo "$(VERSION) is not pushed; run make publish VERSION=$(VERSION) first"; exit 1; }
	@for m in $(ADAPTERS); do \
		(cd $$m && go mod edit -require=$(MODULE)@$(VERSION) && go mod tidy) || exit 1; \
	done
	@git diff --quiet || git commit -am "Point adapters at $(VERSION)"
	@for m in $(ADAPTERS); do \
		git tag -a $$m/$(VERSION) -m "Release $$m $(VERSION)" || exit 1; \
		git push origin $$m/$(VERSION) || exit 1; \
	done

publish-pkg:
	@for i in 1 2 3 4 5 6; do \
		GOPROXY=proxy.golang.org go list -m $(MODULE)@$(VERSION) >/dev/null 2>&1 && exit 0; \
		echo "waiting for the proxy to index $(MODULE)@$(VERSION) ($$i/6)"; \
		sleep 10; \
	done; \
	{ echo "proxy still does not have $(MODULE)@$(VERSION)"; exit 1; }
