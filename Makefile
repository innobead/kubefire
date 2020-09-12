PROJECT := $(shell basename $(CURDIR))
COMMIT := $(shell git rev-parse --short HEAD)-$(shell date "+%Y%m%d%H%M%S")
TAG := $(shell git describe --tags --dirty)
IMAGES := $(shell ./hack/generate-image-info.sh --image)
IMAGES_SUSE := sle15:15.1 sle15:15.2
KERNELS := $(shell ./hack/generate-image-info.sh --kernel)
GOBIN := $(shell go env GOBIN)
GOARCH := $(shell go env GOARCH)

CR_USERNAME := $(CR_USERNAME)
CR_PAT := $(CR_PAT)
CR_PATH ?= ghcr.io/
CR_IMAGE_PREFIX := $(CR_PATH)innobead
KERNEL_IMAGE_NAME=${CR_IMAGE_PREFIX}/$(PROJECT)-ignite-kernel
BUILD_SUSE_IMAGES ?=

ContainerdVersion := v1.4.0
IgniteVersion := v0.7.1
CniVersion := v0.8.6
RuncVersion := v1.0.0-rc92

GO_LINKFLAGS := -X=github.com/innobead/kubefire/internal/config.BuildVersion=$(COMMIT)
GO_LINKFLAGS := -X=github.com/innobead/kubefire/internal/config.TagVersion=$(TAG) $(GO_LINKFLAGS)
GO_LINKFLAGS := -X=github.com/innobead/kubefire/internal/config.ContainerdVersion=$(ContainerdVersion) $(GO_LINKFLAGS)
GO_LINKFLAGS := -X=github.com/innobead/kubefire/internal/config.IgniteVersion=$(IgniteVersion) $(GO_LINKFLAGS)
GO_LINKFLAGS := -X=github.com/innobead/kubefire/internal/config.CniVersion=$(CniVersion) $(GO_LINKFLAGS)
GO_LINKFLAGS := -X=github.com/innobead/kubefire/internal/config.RuncVersion=$(RuncVersion) $(GO_LINKFLAGS)
GO_LDFLAGS := -ldflags "$(GO_LINKFLAGS)"

BUILD_DIR := $(CURDIR)/target
BUILD_CNI_DIR := $(BUILD_DIR)/cni
BUILD_TMP_DIR := $(CURDIR)/.build
BUILD_CACHE_DIR := $(CURDIR)/.cache
BUILD_GENERATE_DIR := $(CURDIR)/generated

.PHONY: help
help:
	@grep -E '^[a-zA-Z%_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: build ## Build and Install executables
	cp $(BUILD_DIR)/kubefire $(GOBIN)

.PHONY: build-all
build-all: clean clean-cni env build build-cni checksum ## Build all

.PHONY: env
env: ## Prepare build env
	 [ -x "$(BUILD_CACHE_DIR)/golangci-lint" ] || (\
			mkdir -p $(BUILD_CACHE_DIR) || true && \
			curl -sfLO https://github.com/golangci/golangci-lint/releases/download/v1.30.0/golangci-lint-1.30.0-linux-amd64.tar.gz && \
			tar -zxvf golangci-lint-1.30.0-linux-amd64.tar.gz && \
			mv ./golangci-lint-1.30.0-linux-amd64/golangci-lint $(BUILD_CACHE_DIR)/ && \
			rm -rf ./golangci-lint-1.30.0-linux-amd64* || true)

.PHONY: build
build: env format ## Build executables (linux/amd64 supported only)
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR) $(GO_LDFLAGS) ./cmd/...

.PHONY: build-cni
build-cni: ## Build CNI executables
	# build `host-local-rev`
	mkdir -p $(BUILD_TMP_DIR) || true
	mkdir -p $(BUILD_CNI_DIR)
	cd $(BUILD_TMP_DIR); \
		TAG=v0.8.6-patch; \
		git clone --branch $${TAG} https://github.com/innobead/plugins; \
        ./plugins/build_linux.sh -ldflags "-extldflags -static -X github.com/containernetworking/plugins/pkg/utils/buildversion.BuildVersion=$${TAG}"; \
		mv ./plugins/bin/host-local $(BUILD_CNI_DIR)/host-local-rev

.PHONY: test
test:  ## Test
	go test -cover -v --tags feature_run_on_ci $(GO_LDFLAGS) ./...

.PHONY: format
format: ## Format source code
	go fmt ./...
	go vet ./...
	go mod tidy
	$(BUILD_CACHE_DIR)/golangci-lint run ./...

.PHONY: checksum
checksum: ## Generate checksum files for built executables
	$(CURDIR)/hack/generate-checksum.sh $(BUILD_DIR)

.PHONY: generate
generate: ## Generate generated files
	$(CURDIR)/hack/generate-image-info.sh --generate

.PHONY: clean
clean: ## Clean build caches
	rm -rf $(BUILD_DIR)
	rm -rf $(BUILD_TMP_DIR)
	rm -rf $(BUILD_CACHE_DIR)

.PHONY: clean-cni
clean-cni: ## CLean build CNI caches
	rm -rf $(BUILD_CNI_DIR)
	rm -rf $(BUILD_TMP_DIR)/plugins

.PHONY: clean-ignite
clean-ignite: ## Clean ignite caches
	$(CURDIR)/hack/clean-ignite.sh

.PHONY: clean-generate
clean-generate: ## Clean generated files
	rm -rf $(BUILD_GENERATE_DIR)

.PHONY: check-publish-image-env
check-publish-image-env: ## Check if the authentication of container registry provided
	([ -n "$(CR_USERNAME)" ] && [ -n "$(CR_PAT)" ]) || (echo "Please setup environment variables (CR_USERNAME, CR_PAT) for publishing images" && exit 1)

.PHONY: build-image-%
build-image-%: ## Build a root image
	docker build --build-arg="RELEASE=$(RELEASE)" -t $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(COMMIT) -f build/images/$*/Dockerfile .
	docker tag $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(COMMIT) $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(RELEASE)

.PHONY: build-images
build-images: ## Build all rootfs images
	for i in $(IMAGES); do $(MAKE) build-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done
ifdef BUILD_SUSE_IMAGES
	for i in $(IMAGES_SUSE); do $(MAKE) build-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done
endif

.PHONY: publish-image-%
publish-image-%: check-publish-image-env build-image-% ## Publish a rootfs image
	echo $(CR_PAT) | docker login $(CR_PATH) -u $(CR_USERNAME) --password-stdin
	docker push $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(COMMIT)
	docker push $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(RELEASE)

.PHONY: publish-images
publish-images: ## Publish rootfs images
	for i in $(IMAGES); do $(MAKE) publish-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done
ifdef BUILD_SUSE_IMAGES
	for i in $(IMAGES_SUSE); do $(MAKE) publish-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done
endif

.PHONY: build-kernels
build-kernels: ## Build kernel images
	for i in $(KERNELS); do $(MAKE) build-kernel-$$i; done

.PHONY: build-kernel-%
build-kernel-%: ## Build a kernel image
	rm -rf ignite/ || true
	git clone git@github.com:weaveworks/ignite.git
	cp build/kernels/config-amd64-$* ignite/images/kernel/enerated && \
 	cd ./ignite/images/kernel && \
 	GOARCH=$(GOARCH) IMAGE_NAME=$(KERNEL_IMAGE_NAME) make build-$*

.PHONY: publish-kernel-%
publish-kernel-%: check-publish-image-env build-kernel-% ## Publish a kernel image
	echo $(CR_PAT) | docker login $(CR_PATH) -u $(CR_USERNAME) --password-stdin
	docker push $(KERNEL_IMAGE_NAME):$*-$(GOARCH)

.PHONY: publish-kernels
publish-kernels: ## Publish all kernel images
	for i in $(KERNELS); do $(MAKE) publish-kernel-$$i; done
