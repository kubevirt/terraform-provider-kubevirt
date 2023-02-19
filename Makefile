GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=kubevirt
ifeq ($(OS),Windows_NT)  # is Windows_NT on XP, 2000, 7, Vista, 10...
	DESTINATION_PREFIX=$(APPDATA)/terraform.d/plugins
else
	DESTINATION_PREFIX=$(HOME)/.terraform.d/plugins
endif

export BIN_DIR=$(CURDIR)/build/_output/bin
export GOROOT=$(BIN_DIR)/go
export GOBIN=$(GOROOT)/bin
export GO=$(GOBIN)/go
export GOFMT=$(GOBIN)/gofmt

all: test install-local

$(GO):
	scripts/install-go.sh $(BIN_DIR)

test-tools: $(GO)
	scripts/install-terraform.sh $(BIN_DIR)

clean: $(GO)
	$(GO) clean
	@echo "==> Removing $(DESTINATION_PREFIX)/$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH) directory"
	@rm -rf $(DESTINATION_PREFIX)/$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)

build: $(GO) test-fmt
	$(GO) build

install-local: build $(GO)
	@mkdir -p $(DESTINATION_PREFIX)/terraform.local/local/kubevirt/1.0.0/$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)
	@cp ./terraform-provider-kubevirt $(DESTINATION_PREFIX)/terraform.local/local/kubevirt/1.0.0/$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)
	@echo "==> Installing plugin to $(DESTINATION_PREFIX)/terraform.local/local/kubevirt/1.0.0/$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)"

test-fmt:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

test-vet: $(GO)
	@echo "go vet ."
	$(GO) vet $$($(GO) list ./... | grep -v vendor/)
	@if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt: $(GO)
	$(GOFMT) -w $(GOFMT_FILES)

test: $(GO) test-fmt
	$(GO) test ./kubevirt/... $(TESTARGS) -timeout=30s -parallel=4

functest: test-tools
	$(CURDIR)/scripts/func-test.sh $(BIN_DIR)

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

cluster-up:
	./kubevirtci up

cluster-down:
	./kubevirtci down

.PHONY: \
	clean \
	build \
	install-local \
	test-fmt \
	test-vet \
	fmt \
	test \
	test-acc \
	functest \
	errcheck \
	test-compile \
	cluster-up \
	cluster-down \


