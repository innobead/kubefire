
CWD=$(shell basename $(CURDIR))
COMMIT=$(shell git rev-parse --short HEAD)

GO_LDFLAGS=-ldflags "-X=github.com/innobead/kubefire/internal/config.BuildVersion=$(COMMIT)"
BUILD_DIR=target

.PHONY: build
build: clean format
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR) $(GO_LDFLAGS) ./cmd/...

.PHONY: format
format:
	go fmt ./...
	go vet ./...
	golangci-lint run ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: build-image-%
build-image-%:
	docker build -t innobead/$(CWD):$*-$(COMMIT) build/images/$*
	docker tag innobead/$(CWD):$*-$(COMMIT) innobead/$(CWD):$*-latest

.PHONY: publish-image-%
publish-image-%: build-image-%
	docker push innobead/$(CWD):$*-$(COMMIT)
	docker push innobead/$(CWD):$*-latest

.PHONY: build-kernel-%
build-kernel-%:
	:

.PHONY: publish-kernel-%
publish-kernel-%: build-kernel-%
	: