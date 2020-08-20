.PHONY: env \
		build-all \
		build \
		build-cni \
		format \
		checksum \
		clean \
		clean-cni \
		clean-ignite-runtime \
		build-images \
		build-image-% \
		publish-images \
		build-kernels \
		build-kernel-% \
		publish-kernels \
		publish-kernel-%

PROJECT:=$(shell basename $(CURDIR))
COMMIT:=$(shell git rev-parse --short HEAD)-$(shell date "+%Y%m%d%H%M%S")
TAG:=$(shell git name-rev --tags --name-only $$(git rev-parse HEAD) | sed s/undefined/master/) # $(shell git describe --tags --dirty)
IMAGES:=centos:8 ubuntu:18.04 ubuntu:20.10 opensuse-leap:15.1 sle15:15.1 opensuse-leap:15.2 sle15:15.2
KERNELS:=$(shell ls ./build/kernels | sed 's/config-amd64-//; /README.md/d;')
GOBIN:=$(shell go env GOBIN)

ContainerdVersion := v1.3.7
IgniteVersion := v0.7.1
CniVersion := v0.8.6
RuncVersion := v1.0.0-rc92

GO_LINKFLAGS:=-X=github.com/innobead/kubefire/internal/config.BuildVersion=$(COMMIT)
GO_LINKFLAGS:=-X=github.com/innobead/kubefire/internal/config.TagVersion=$(TAG) $(GO_LINKFLAGS)
GO_LINKFLAGS:=-X=github.com/innobead/kubefire/internal/config.ContainerdVersion=$(ContainerdVersion) $(GO_LINKFLAGS)
GO_LINKFLAGS:=-X=github.com/innobead/kubefire/internal/config.IgniteVersion=$(IgniteVersion) $(GO_LINKFLAGS)
GO_LINKFLAGS:=-X=github.com/innobead/kubefire/internal/config.CniVersion=$(CniVersion) $(GO_LINKFLAGS)
GO_LINKFLAGS:=-X=github.com/innobead/kubefire/internal/config.RuncVersion=$(RuncVersion) $(GO_LINKFLAGS)
GO_LDFLAGS:=-ldflags "$(GO_LINKFLAGS)"

BUILD_DIR:=$(CURDIR)/target
BUILD_CNI_DIR:=$(BUILD_DIR)/cni
BUILD_TMP_DIR:=$(CURDIR)/.build
BUILD_CACHE_DIR:=$(CURDIR)/.cache

help:
	@grep -E '^[a-zA-Z%_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install: build ## Build and Install executables
	cp $(BUILD_DIR)/kubefire $(GOBIN)

build-all: clean clean-cni env build build-cni checksum ## Build all

env: ## Prepare build env
	 [ ! -x "$(BUILD_CACHE_DIR)/golangci-lint" ] && \
			mkdir -p $(BUILD_CACHE_DIR) && \
			curl -sfLO https://github.com/golangci/golangci-lint/releases/download/v1.30.0/golangci-lint-1.30.0-linux-amd64.tar.gz && \
			tar -zxvf golangci-lint-1.30.0-linux-amd64.tar.gz && \
			mv ./golangci-lint-1.30.0-linux-amd64/golangci-lint $(BUILD_CACHE_DIR)/ && \
			rm -rf ./golangci-lint-1.30.0-linux-amd64* || true

build: format ## Build executables (linux/amd64 supported only)
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR) $(GO_LDFLAGS) ./cmd/...

build-cni: ## Build CNI executables
	# build `host-local-rev`
	mkdir -p $(BUILD_TMP_DIR) || true
	mkdir -p $(BUILD_CNI_DIR)
	cd $(BUILD_TMP_DIR); \
		TAG=v0.8.6-patch; \
		git clone --branch $${TAG} https://github.com/innobead/plugins; \
        ./plugins/build_linux.sh -ldflags "-extldflags -static -X github.com/containernetworking/plugins/pkg/utils/buildversion.BuildVersion=$${TAG}"; \
		mv ./plugins/bin/host-local $(BUILD_CNI_DIR)/host-local-rev

format: ## Format source code
	go fmt ./...
	go vet ./...
	go mod tidy
	$(BUILD_CACHE_DIR)/golangci-lint run ./...

checksum: ## Generate checksum files for built executables
	$(CURDIR)/hack/generate-checksum.sh $(BUILD_DIR)

clean: ## Clean build caches
	rm -rf $(BUILD_DIR)
	rm -rf $(BUILD_TMP_DIR)
	rm -rf $(BUILD_CACHE_DIR)

clean-cni: ## CLean build CNI caches
	rm -rf $(BUILD_CNI_DIR)
	rm -rf $(BUILD_TMP_DIR)/plugins

clean-ignite: ## Clean ignite caches
	$(CURDIR)/hack/clean-ignite.sh

build-image-%: ## Build a root image
	docker build --build-arg="RELEASE=$(RELEASE)" -t innobead/$(PROJECT)-$*:$(COMMIT) -f build/images/$*/Dockerfile .
	docker tag innobead/$(PROJECT)-$*:$(COMMIT) innobead/$(PROJECT)-$*:$(RELEASE)

build-images: ## Build all rootfs images
	for i in $(IMAGES); do $(MAKE) build-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done

publish-image-%: build-image-% ## Publish a rootfs image
	docker push innobead/$(PROJECT)-$*:$(COMMIT)
	docker push innobead/$(PROJECT)-$*:$(RELEASE)

publish-images: ## Publish rootfs images
	for i in $(IMAGES); do $(MAKE) publish-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done

build-kernels: ## Build kernel images
	for i in $(KERNELS); do $(MAKE) build-kernel-$$i; done

build-kernel-%: ## Build a kernel image
	git clone git@github.com:weaveworks/ignite.git
	cp build/kernels/config-amd64-$* ignite/images/kernel/enerated && \
 	cd ./ignite/images/kernel && \
 	make build-$*

publish-kernel-%: build-kernel-% ## Publish a kernel image
	docker push innobead/$(PROJECT)-kernel-$*:$(COMMIT)
	docker push innobead/$(PROJECT)-kernel-$*:latest

publish-kernels: ## Publish all kernel images
	for i in $(KERNELS); do $(MAKE) publish-kernel-$$i; done
