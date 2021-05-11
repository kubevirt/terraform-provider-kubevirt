TEST?=$$(go list ./... | grep -v 'vendor' | grep -v ci-tests)
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=kubevirt

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
ifeq ($(OS),Windows_NT)  # is Windows_NT on XP, 2000, 7, Vista, 10...
	DESTINATION=$(APPDATA)/terraform.d/plugins/$(GOOS)_$(GOARCH)
else
	DESTINATION=$(HOME)/.terraform.d/plugins/$(GOOS)_$(GOARCH)
endif

all: test install

clean:
	go clean
	@echo "==> Removing $(DESTINATION) directory"
	@rm -rf $(DESTINATION)

build: test-fmt
	go build

install: build
	@echo "==> Installing plugin to $(DESTINATION)"
	@mkdir -p $(DESTINATION)
	@cp ./terraform-provider-kubevirt $(DESTINATION)

test-fmt:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

test-vet:
	@echo "go vet ."
	go vet $$(go list ./... | grep -v vendor/)
	@if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

test: test-fmt
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

test-acc: test-fmt
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

functest:
	export KUBECONFIG=$(sh -c "cluster-up/kubeconfig.sh")
	go test ./ci-tests/... -timeout 99999s

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

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


