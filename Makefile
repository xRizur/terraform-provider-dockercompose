BINARY   = terraform-provider-dockercompose
VERSION  = 1.2.0
OS      ?= $(shell go env GOOS)
ARCH    ?= $(shell go env GOARCH)

# Terraform plugin cache dir for dev_overrides
PLUGIN_DIR = $(HOME)/.terraform.d/plugins/registry.terraform.io/xrizur/dockercompose/$(VERSION)/$(OS)_$(ARCH)

.PHONY: build install clean

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY) $(PLUGIN_DIR)/$(BINARY)_v$(VERSION)

clean:
	rm -f $(BINARY)
