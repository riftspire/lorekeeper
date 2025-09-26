APP_NAME = "lorekeeper"

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsibile for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mUsage:\n  \033[0mmake \033[36m<target>\033[0m\n\n\033[1;36mTargets:\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "    \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n  \033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: setup
setup: ## Setup the computer.
	echo testing

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: mod
mod: ## Run go mod commands to ensure dependencies are up to date.
	go mod download
	go mod tidy
	go mod verify

.PHONY: test
test: fmt vet ## Run tests.
	go test -v $$(go list ./...)\
	 -coverprofile cover.out

.PHONY: test-verbose
test-verbose: fmt vet ## Run tests with verbose output.
	go test -v $$(go list ./...)\
	 -coverprofile cover.out\
	 -ginkgo.v\
	 -ginkgo.show-node-events\
	 -ginkgo.randomize-all

.PHONY: gopls-check
gopls-check: gopls ## Run gopls check to ensure code is formatted and analyzed.
	$(GOPLS) check -severity=hint $$(find . -name "*.go")

.PHONY: lint
lint: fmt vet golangci-lint gopls-check ## Run golangci-lint and gopls check.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: fmt vet golangci-lint ## Run golangci-lint linter and perform fixes.
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: docs
docs: ## Generate the docs using Go templates.
	go run ./cmd/$(APP_NAME)/. --docs

.PHONY: build
build: ## Build a binary from the Go code.
	go build\
	 -o ./bin/$(APP_NAME) ./cmd/$(APP_NAME)


##@ Dependencies - Dev

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
GOPLS = $(LOCALBIN)/gopls-$(GOPLS_VERSION)

## Tool Versions
GOLANGCI_LINT_VERSION ?= v2.1.6
GOPLS_VERSION ?= v0.19.1

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: gopls
gopls: $(GOPLS) ## Download gopls locally if necessary.
$(GOPLS): $(LOCALBIN)
	$(call go-install-tool,$(GOPLS),golang.org/x/tools/gopls,$(GOPLS_VERSION))

# go-install-tool will 'go install' any package with custom target and name of
# binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef
