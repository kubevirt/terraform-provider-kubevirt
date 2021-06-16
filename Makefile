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

all: test install

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

install: build $(GO)
	@echo "==> Installing plugin to $(DESTINATION_PREFIX)/$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)"
	@mkdir -p $(DESTINATION_PREFIX)/$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)
	@cp ./terraform-provider-kubevirt $(DESTINATION_PREFIX)/$(shell $(GO) env GOOS)_$(shell $(GO) env GOARCH)

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
	export PATH=$(BIN_DIR)/$(PATH)
	@sh -c "'$(CURDIR)/scripts/func-test.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

cluster-up:
	sh -c "./cluster-up/up.sh" 
	export KUBECONFIG=$(sh -c "cluster-up/kubeconfig.sh")
	sh -c "./cluster-kubevirt-deploy/kubevirt-deploy.sh"

cluster-down:
	sh -c "./cluster-up/down.sh" 

.PHONY: \
	clean \
	build \
	install \
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


