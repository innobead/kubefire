.PHONY: build \
		format \
		clean \
		clean-ignite-runtime \
		build-images \
		build-image-% \
		publish-images \
		publish-image-% \
		build-kernels \
		build-kernel-% \
		publish-kernels \
		publish-kernel-%

CWD=$(shell basename $(CURDIR))
COMMIT=$(shell git rev-parse --short HEAD)
TAG=$(shell git name-rev --tags --name-only $$(git rev-parse HEAD) | sed s/undefined/master/)
IMAGES=$(shell ls ./build/images)
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

clean-ignite-runtime:
	@sudo ignite rm -f $(sudo ignite ps -aq) || true
	@sudo ignite rmi $(sudo ignite images ls | awk '{print $$1}' | sed '1d') || true
	@sudo ignite rmk $(sudo ignite kernels ls | awk '{print $$1}' | sed '1d') || true
	@ctr -n firecracker i rm $(ctr -n firecracker images ls | awk '{print $$1}' | sed '1d' )

build-images:
	for i in $(IMAGES); do $(MAKE) build-image-$$i; done

build-image-%:
	docker build -t innobead/$(CWD)-$*:$(COMMIT) -f build/images/$*/Dockerfile .
	docker tag innobead/$(CWD)-$*:$(COMMIT) innobead/$(CWD)-$*:latest

publish-image-%: build-image-%
	docker push innobead/$(CWD)-$*:$(COMMIT)
	docker push innobead/$(CWD)-$*:latest

publish-images:
	for i in $(IMAGES); do $(MAKE) publish-image-$$i; done

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

