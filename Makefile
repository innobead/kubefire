.PHONY: build clean images build-image-% publish-image-%

CWD=$(shell basename $(CURDIR))
COMMIT=$(shell git rev-parse --short HEAD)

GO_LDFLAGS=-ldflags "-X=github.com/innobead/kubefire/internal/config.BuildVersion=$(COMMIT)"
BUILD_DIR=target

build: clean
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR) $(GO_LDFLAGS) ./cmd/...

clean:
	rm -rf $(BUILD_DIR)

build-image-%:
	docker build -t innobead/$(CWD):$*-$(COMMIT) build/images/$*
	docker tag innobead/$(CWD):$*-$(COMMIT) innobead/$(CWD):$*-latest

publish-image-%: build-image-%
	docker push innobead/$(CWD):$*-$(COMMIT)
	docker push innobead/$(CWD):$*-latest
