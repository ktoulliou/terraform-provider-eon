BINARY=terraform-provider-eon
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)
PLUGIN_DIR=~/.terraform.d/plugins/registry.terraform.io/ktoulliou/eon/0.1.0/$(OS_ARCH)

.PHONY: build install clean

build:
	go build -o $(BINARY)

install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY) $(PLUGIN_DIR)/

clean:
	rm -f $(BINARY)
