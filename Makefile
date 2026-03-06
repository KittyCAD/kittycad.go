NAME := kittycad

# If this session isn't interactive, then we don't want to allocate a
# TTY, which would fail, but if it is interactive, we do want to attach
# so that the user can send e.g. ^C through.
INTERACTIVE := $(shell [ -t 0 ] && echo 1 || echo 0)
ifeq ($(INTERACTIVE), 1)
	DOCKER_FLAGS += -t
endif

# Set our default go compiler
GO := go
GO_BIN_DIR := $(shell gobin="$$( $(GO) env GOBIN )"; if [ -n "$$gobin" ]; then printf "%s" "$$gobin"; else printf "%s/bin" "$$( $(GO) env GOPATH )"; fi)
GOIMPORTS := $(or $(shell command -v goimports 2>/dev/null),$(GO_BIN_DIR)/goimports)
GOLINT := $(or $(shell command -v golint 2>/dev/null),$(GO_BIN_DIR)/golint)
STATICCHECK := $(or $(shell command -v staticcheck 2>/dev/null),$(GO_BIN_DIR)/staticcheck)

VERSION := $(shell cat $(CURDIR)/VERSION.txt)

.PHONY: generate
generate:
	@# Ensure goimports is available, but avoid network if already installed
	@[ -x "$(GOIMPORTS)" ] || $(GO) install golang.org/x/tools/cmd/goimports@latest
	@# Build the code generator
	go build -o $(CURDIR)/generate $(CURDIR)/cmd
	./generate
	$(GOIMPORTS) -w *.go
	gofmt -s -w *.go
	go mod tidy

.PHONY: build
build: $(NAME) ## Builds a dynamic package.

$(NAME): $(wildcard *.go) $(wildcard */*.go)
	@echo "+ $@"
	$(GO) build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) .

all: generate clean build fmt lint test staticcheck vet install ## Runs a clean, build, fmt, lint, test, staticcheck, vet and install.

.PHONY: fmt
fmt: ## Verifies all files have been `gofmt`ed.
	@echo "+ $@"
	@out="$$(gofmt -s -l . | grep -v '.pb.go:' | grep -v '.twirp.go:' | grep -v vendor)"; \
	if [ -n "$$out" ]; then \
		printf '%s\n' "$$out" >&2; \
		exit 1; \
	fi

.PHONY: lint
lint: ## Verifies `golint` passes.
	@echo "+ $@"
	@[ -x "$(GOLINT)" ] || $(GO) install golang.org/x/lint/golint@latest
	@out="$$( $(GOLINT) ./... | grep -v '.pb.go:' | grep -v '.twirp.go:' | grep -v vendor )"; \
	if [ -n "$$out" ]; then \
		printf '%s\n' "$$out" >&2; \
		exit 1; \
	fi

.PHONY: test
test: ## Runs the go tests.
	@echo "+ $@"
	@$(GO) test -v -tags "$(BUILDTAGS) cgo" $(shell $(GO) list ./... | grep -v vendor)

.PHONY: vet
vet: ## Verifies `go vet` passes.
	@echo "+ $@"
	@$(GO) vet $(shell $(GO) list ./... | grep -v vendor)

.PHONY: staticcheck
staticcheck: ## Verifies `staticcheck` passes.
	@echo "+ $@"
	@[ -x "$(STATICCHECK)" ] || $(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	@$(STATICCHECK) $(shell $(GO) list ./... | grep -v vendor | grep -v "src\/internal" | grep -v "src\/hash")

.PHONY: cover
cover: ## Runs go test with coverage.
	@echo "" > coverage.txt
	@for d in $(shell $(GO) list ./... | grep -v vendor); do \
		$(GO) test -race -coverprofile=profile.out -covermode=atomic "$$d"; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done;

.PHONY: install
install: ## Installs the executable or package.
	@echo "+ $@"
	$(GO) install -a -tags "$(BUILDTAGS)" ${GO_LDFLAGS} .

.PHONY: clean
clean: ## Cleanup any build binaries or packages.
	@echo "+ $@"
	$(RM) $(NAME)
	$(RM) -r $(BUILDDIR)

.PHONY: tag
tag: ## Create a new git tag to prepare to build a release.
	git tag -sa $(VERSION) -m "$(VERSION)"
	@echo "Run git push origin $(VERSION) to push your new tag to GitHub and trigger a release."

.PHONY: bump-version
BUMP := patch
bump-version: ## Bump the version in the version file. Set BUMP to [ patch | major | minor ].
	@$(GO) get -u github.com/jessfraz/junk/sembump || true # update sembump tool
	$(eval NEW_VERSION = $(shell sembump --kind $(BUMP) $(VERSION)))
	@echo "Bumping VERSION.txt from $(VERSION) to $(NEW_VERSION)"
	echo $(NEW_VERSION) > VERSION.txt
	@echo "Updating links to download binaries in README.md"
	sed -i s/$(VERSION)/$(NEW_VERSION)/g README.md
	git add VERSION.txt README.md
	git commit -vsam "Bump version to $(NEW_VERSION)"
	@echo "Run make tag to create and push the tag for new version $(NEW_VERSION)"

.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
