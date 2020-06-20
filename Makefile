.PHONY: build \
		clean \
		format \
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

GO_LDFLAGS=-ldflags "-X=github.com/innobead/kubefire/internal/config.BuildVersion=$(COMMIT) -X=github.com/innobead/kubefire/internal/config.TagVersion=$(TAG)"
BUILD_DIR=target

build: clean format
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR) $(GO_LDFLAGS) ./cmd/...

format:
	go fmt ./...
	go vet ./...
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

build-images:
	for i in $(IMAGES); do $(MAKE) build-image-$$i; done

build-image-%:
	docker build -t innobead/$(CWD):$*-$(COMMIT) build/images/$*
	docker tag innobead/$(CWD):$*-$(COMMIT) innobead/$(CWD):$*-latest

publish-images:
	for i in $(IMAGES); do $(MAKE) publish-image-$$i; done

publish-image-%: build-image-%
	docker push innobead/$(CWD):$*-$(COMMIT)
	docker push innobead/$(CWD):$*-latest

build-kernel-%:
	:

publish-kernel-%: build-kernel-%
	: