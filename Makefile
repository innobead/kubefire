PROJECT := $(shell basename $(CURDIR))
COMMIT := $(shell git rev-parse --short HEAD)-$(shell date "+%Y%m%d%H%M%S")
TAG := $(shell git describe --tags --dirty)
IMAGES := $(shell ./hack/generate-image-info.sh --image)
IMAGES_SUSE := sle15:15.3
KERNELS := $(shell ./hack/generate-image-info.sh --kernel)
GOBIN := $(or $(shell go env GOBIN), $(HOME)/go/bin)
GOARCH := $(shell go env GOARCH)
DOCKER_BUILDX := docker buildx build --platform linux/arm64,linux/amd64

CR_USERNAME := $(CR_USERNAME)
CR_PAT := $(CR_PAT)
CR_PATH ?= ghcr.io/
CR_IMAGE_PREFIX := $(CR_PATH)innobead
KERNEL_IMAGE_NAME=${CR_IMAGE_PREFIX}/$(PROJECT)-ignite-kernel
BUILD_SUSE_IMAGES ?=

GolangCILintVersion := 1.32.2

ContainerdVersion := v1.4.4
IgniteVersion := v0.9.0
CniVersion := v0.9.1
RuncVersion := v1.0.0-rc93

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

BUILDX_BUILDER := arch_builder
HAS_BUILDX_BUILDER := $(shell docker buildx ls | grep $(BUILDX_BUILDER))

.PHONY: help
help:
	@grep -E '^[a-zA-Z%_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: install
install: build ## Build and Install executables
	mkdir -p $(GOBIN) || true
	cp $(BUILD_DIR)/kubefire-linux-$(GOARCH) $(GOBIN)/kubefire

.PHONY: build-all
build-all: clean clean-cni env build build-cni checksum ## Build all

.PHONY: env
env: ## Prepare build env
	 [ -x "$(BUILD_CACHE_DIR)/golangci-lint" ] || (\
			mkdir -p $(BUILD_CACHE_DIR) || true && \
			curl -sfLO https://github.com/golangci/golangci-lint/releases/download/v${GolangCILintVersion}/golangci-lint-$(GolangCILintVersion)-linux-$(GOARCH).tar.gz && \
			tar -zxvf golangci-lint-$(GolangCILintVersion)-linux-$(GOARCH).tar.gz && \
			mv ./golangci-lint-$(GolangCILintVersion)-linux-$(GOARCH)/golangci-lint $(BUILD_CACHE_DIR)/ && \
			rm -rf ./golangci-lint-$(GolangCILintVersion)-linux* || true)

.PHONY: build
build: env format ## Build executables (linux/amd64 supported only)
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/kubefire-linux-amd64 $(GO_LDFLAGS) ./cmd/kubefire
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/kubefire-linux-arm64 $(GO_LDFLAGS) ./cmd/kubefire

.PHONY: build-cni
build-cni: ## Build CNI executables
	# build `host-local-rev`
	mkdir -p $(BUILD_TMP_DIR) || true
	mkdir -p $(BUILD_CNI_DIR)
	cd $(BUILD_TMP_DIR); \
		TAG=$(CniVersion)-patch; \
		git clone --branch $${TAG} https://github.com/innobead/plugins; \
        GOOS=linux GOARCH=amd64 ./plugins/build_linux.sh -ldflags "-extldflags -static -X github.com/containernetworking/plugins/pkg/utils/buildversion.BuildVersion=$${TAG}"; \
		mv ./plugins/bin/host-local $(BUILD_CNI_DIR)/host-local-rev-linux-amd64; \
        GOOS=linux GOARCH=arm64 ./plugins/build_linux.sh -ldflags "-extldflags -static -X github.com/containernetworking/plugins/pkg/utils/buildversion.BuildVersion=$${TAG}"; \
		mv ./plugins/bin/host-local $(BUILD_CNI_DIR)/host-local-rev-linux-arm64

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
build-image-%: init-buildx ## Build a root image
	$(DOCKER_BUILDX) \
		--build-arg="RELEASE=$(RELEASE)" \
		--output "type=image,push=false" \
		-t $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(COMMIT) -t $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(RELEASE) \
		-f build/images/$*/Dockerfile .

.PHONY: build-images
build-images: ## Build all rootfs images
	for i in $(IMAGES); do $(MAKE) build-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done
ifdef BUILD_SUSE_IMAGES
	for i in $(IMAGES_SUSE); do $(MAKE) build-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done
endif

.PHONY: publish-image-%
publish-image-%: check-publish-image-env init-buildx ## Publish a rootfs image
	echo $(CR_PAT) | docker login $(CR_PATH) -u $(CR_USERNAME) --password-stdin
	$(DOCKER_BUILDX) \
		--build-arg="RELEASE=$(RELEASE)" \
		--output "type=image,push=true" \
		-t $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(COMMIT) -t $(CR_IMAGE_PREFIX)/$(PROJECT)-$*:$(RELEASE) \
		-f build/images/$*/Dockerfile .

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
	git clone https://github.com/weaveworks/ignite.git
	cp build/kernels/config-* ignite/images/kernel/generated/ && \
 	cd ./ignite/images/kernel && \
    docker build -t $(KERNEL_IMAGE_NAME):$*-amd64 \
		--build-arg KERNEL_VERSION=$* \
		--build-arg ARCH=x86 \
		--build-arg GOARCH=amd64 \
		--build-arg ARCH_MAKE_PARAMS=${ARCH_MAKE_PARAMS} \
		--build-arg VMLINUX_PATH=vmlinux . && \
    docker build -t $(KERNEL_IMAGE_NAME):$*-arm64 \
		--build-arg KERNEL_VERSION=$* \
		--build-arg ARCH=arm64 \
		--build-arg GOARCH=arm64 \
		--build-arg ARCH_MAKE_PARAMS="ARCH=arm64 CROSS_COMPILE=aarch64-linux-gnu-" \
		--build-arg VMLINUX_PATH=arch/arm64/boot/Image .

.PHONY: publish-kernel-%
publish-kernel-%: check-publish-image-env build-kernel-% ## Publish a kernel image
	echo $(CR_PAT) | docker login $(CR_PATH) -u $(CR_USERNAME) --password-stdin
	cd ./ignite/images/kernel && \
	../../hack/push-manifest-list.sh $(KERNEL_IMAGE_NAME):$* amd64 arm64

.PHONY: publish-kernels
publish-kernels: ## Publish all kernel images
	for i in $(KERNELS); do $(MAKE) publish-kernel-$$i; done

.PHONY: init-buildx
init-buildx: ## Init multiple arch builder
ifeq ($(strip $(HAS_BUILDX_BUILDER)),)
	# FIXME: https://github.com/moby/buildkit/pull/1636, use the buildkit from master branch to resolve 401 unathorized issue. It should be fixed after 0.7.2.
	docker buildx create --name $(BUILDX_BUILDER) --driver-opt image=moby/buildkit:master
	docker buildx inspect $(BUILDX_BUILDER) --bootstrap
endif
	docker buildx use $(BUILDX_BUILDER)

.PHONY: uninit-buildx
uninit-buildx: ## Uninit multiple arch builder
ifneq ($(strip $(HAS_BUILDX_BUILDER)),)
	docker buildx rm $(BUILDX_BUILDER)
endif