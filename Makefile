.PHONY: build vet test race cover fmt fmt-check tidy check standalone \
	version-guard publish publish-adapters publish-pkg

MODULE   := github.com/MaksMakarskyi/dam
ADAPTERS := damecho damgin
LIBS     := . $(ADAPTERS)
MODULES  := $(LIBS) _example

# Release inputs, set on the command line:
#
#	make publish VERSION=v0.2.0
#	make publish-adapters VERSION=v0.1.0 CORE_VERSION=v0.2.0
#
# CORE_VERSION is the root version the adapters require. It matches VERSION
# only while the adapters release in lockstep with the core.
VERSION      ?=
CORE_VERSION ?= $(VERSION)

# each runs a shell command once per module directory, stopping at the first
# failure. Every nested module needs its own visit: ./... stops at a module
# boundary, so the root never reaches the adapters or the examples.
define each
@for m in $(1); do \
	(cd $$m && $(2)) || exit 1; \
done
endef

# ---------------------------------------------------------------- development

build:
	$(call each,$(MODULES),go build ./...)

vet:
	$(call each,$(MODULES),go vet ./...)

test:
	$(call each,$(LIBS),go test -shuffle=on ./...)

race:
	$(call each,$(LIBS),go test -race -shuffle=on -count=1 ./...)

cover:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1
	@go tool cover -html=coverage.out -o coverage.html

fmt:
	@gofmt -l -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || { echo "unformatted:"; gofmt -l .; exit 1; }

tidy:
	$(call each,$(MODULES),go mod tidy)

# Builds each adapter the way a consumer sees it: no replace, so the required
# root version has to be published and has to contain the API the adapter uses.
# Every other target reads the core off disk, which hides this entirely.
standalone:
	@tmp=$$(mktemp -d) && \
	for m in $(ADAPTERS); do \
		mkdir -p $$tmp/$$m && cp $$m/*.go $$m/go.mod $$m/go.sum $$tmp/$$m/ 2>/dev/null; \
		(cd $$tmp/$$m && go mod edit -dropreplace=$(MODULE) \
			&& GOFLAGS=-mod=mod go build ./...) \
			|| { rm -rf $$tmp; echo "$$m does not build against the published $(MODULE)"; exit 1; }; \
	done; \
	rm -rf $$tmp; \
	echo "adapters build against the published $(MODULE)"

check: fmt-check build vet race

# -------------------------------------------------------------------- release

version-guard:
	@test -n "$(VERSION)" || { echo "usage: make $(MAKECMDGOALS) VERSION=vX.Y.Z"; exit 1; }
	@echo "$(VERSION)" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$$' \
		|| { echo "VERSION must look like v1.2.3, got: $(VERSION)"; exit 1; }
	@git diff --quiet HEAD || { echo "working tree is dirty; commit first"; exit 1; }

# Tags and publishes the root module. Run before publish-adapters: the adapters
# require the root at CORE_VERSION, which has to exist first.
publish: version-guard check
	@! git rev-parse -q --verify refs/tags/$(VERSION) >/dev/null \
		|| { echo "tag $(VERSION) already exists"; exit 1; }
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@$(MAKE) publish-pkg MODULE=$(MODULE) VERSION=$(VERSION)

# Points the adapters at the published root, then tags each under its own path
# prefix, which is how a nested module is versioned.
publish-adapters: version-guard check
	@echo "$(CORE_VERSION)" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$$' \
		|| { echo "CORE_VERSION must look like v1.2.3, got: $(CORE_VERSION)"; exit 1; }
	@for m in $(ADAPTERS); do \
		! git rev-parse -q --verify refs/tags/$$m/$(VERSION) >/dev/null \
			|| { echo "tag $$m/$(VERSION) already exists"; exit 1; }; \
	done
	$(call each,$(ADAPTERS),go mod edit -require=$(MODULE)@$(CORE_VERSION) && go mod tidy)
	@$(MAKE) standalone
	@git diff --quiet || git commit -am "Point adapters at $(MODULE) $(CORE_VERSION)"
	@for m in $(ADAPTERS); do \
		git tag -a $$m/$(VERSION) -m "Release $$m $(VERSION)" || exit 1; \
		git push origin $$m/$(VERSION) || exit 1; \
	done
	@for m in $(ADAPTERS); do \
		$(MAKE) publish-pkg MODULE=$(MODULE)/$$m VERSION=$(VERSION) || exit 1; \
	done

publish-pkg:
	@for i in 1 2 3 4 5 6; do \
		GOPROXY=proxy.golang.org go list -m $(MODULE)@$(VERSION) >/dev/null 2>&1 && exit 0; \
		echo "waiting for the proxy to index $(MODULE)@$(VERSION) ($$i/6)"; \
		sleep 10; \
	done; \
	{ echo "proxy still does not have $(MODULE)@$(VERSION)"; exit 1; }
