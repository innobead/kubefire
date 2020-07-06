.PHONY: build \
		format \
		clean \
		clean-ignite-runtime \
		build-images \
		build-image-% \
		publish-images \
		build-kernels \
		build-kernel-% \
		publish-kernels \
		publish-kernel-%

CWD=$(shell basename $(CURDIR))
COMMIT=$(shell git rev-parse --short HEAD)
TAG=$(shell git name-rev --tags --name-only $$(git rev-parse HEAD) | sed s/undefined/master/)
IMAGES=centos:8 ubuntu:18.04 ubuntu:20.10 opensuse-leap:15.1 sle15:15.1
KERNELS=$(shell ls ./build/kernels)
GOBIN=$(shell go env GOBIN)

GO_LDFLAGS=-ldflags "-X=github.com/innobead/kubefire/internal/config.BuildVersion=$(COMMIT) -X=github.com/innobead/kubefire/internal/config.TagVersion=$(TAG)"
BUILD_DIR=target

install: build
	cp $(BUILD_DIR)/kubefire $(GOBIN)

build: clean format
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR) $(GO_LDFLAGS) ./cmd/...

format:
	go fmt ./...
	go vet ./...
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

# Clean ignite caches
clean-ignite:
	sudo ignite rm -f $$(sudo ignite ps -aq) &>/dev/null || echo "> No VMs to delete from ignite"
	sudo ignite rmi $$(sudo ignite images ls | awk '{print $$1}' | sed '1d') &>/dev/null || echo "> No images to delete from ignite"
	sudo ignite rmk $$(sudo ignite kernels ls | awk '{print $$1}' | sed '1d') &>/dev/null || echo "> No kernels to delete from ignite"
	sudo ctr -n firecracker i rm $$(sudo ctr -n firecracker images ls | awk '{print $$1}' | sed '1d') &>/dev/null

build-image-%:
	docker build --build-arg="RELEASE=$(RELEASE)" -t innobead/$(CWD)-$*:$(COMMIT) -f build/images/$*/Dockerfile .
	docker tag innobead/$(CWD)-$*:$(COMMIT) innobead/$(CWD)-$*:$(RELEASE)

build-images:
	for i in $(IMAGES); do $(MAKE) build-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done

publish-image-%: build-image-%
	docker push innobead/$(CWD)-$*:$(COMMIT)
	docker push innobead/$(CWD)-$*:$(RELEASE)

publish-images:
	for i in $(IMAGES); do $(MAKE) publish-image-$$(echo $$i | awk -F: '{print $$1}') RELEASE=$$(echo $$i | awk -F: '{print $$2}'); done

build-kernel-%:
	git clone git@github.com:weaveworks/ignite.git
	cp build/kernels/config-amd64-$* ignite/images/kernel/enerated && \
 	cd ./ignite/images/kernel && \
 	make build-$*

publish-kernel-%: build-kernel-%
	docker push innobead/$(CWD)-kernel-$*:$(COMMIT)
	docker push innobead/$(CWD)-kernel-$*:latest

publish-kernels:
	for i in $(KERNELS); do $(MAKE) publish-kernel-$$i; done

