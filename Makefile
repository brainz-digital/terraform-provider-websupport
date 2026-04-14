BINARY    := terraform-provider-websupport
VERSION   := dev
INSTALL_DIR := $(HOME)/.terraform.d/plugins/registry.terraform.io/brainz-digital/websupport/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)

.PHONY: tidy build install dev clean

tidy:
	go mod tidy

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(INSTALL_DIR)
	mv $(BINARY) $(INSTALL_DIR)/$(BINARY)_v$(VERSION)
	@echo "Installed to $(INSTALL_DIR)"

# `dev` writes a ~/.terraformrc that points at the local build via dev_overrides.
# After running `make dev`, terraform/terragrunt will use the local binary
# without needing any provider source/version block in HCL.
dev: build
	@echo "Add this to ~/.terraformrc:"
	@echo
	@echo 'provider_installation {'
	@echo '  dev_overrides {'
	@printf '    "registry.terraform.io/brainz-digital/websupport" = "%s"\n' "$(PWD)"
	@echo '  }'
	@echo '  direct {}'
	@echo '}'

clean:
	rm -f $(BINARY)
