GIT_COMMIT?=$(shell git rev-parse HEAD)
GIT_COMMIT_SHORT?=$(shell git rev-parse --short HEAD)
GIT_TAG?=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "v0.0.0" )
TAG?=${GIT_TAG}-${GIT_COMMIT_SHORT}
REPO?=quay.io/costoolkit/upgradechannel-discovery
export TEST_IMAGE?=$(REPO):$(TAG)
export ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
export ROS_CHART?=$(shell find $(ROOT_DIR) -type f  -name "rancheros-operator*.tgz" -print)
KUBE_VERSION?="v1.22.7"
export CLUSTER_NAME?=upgradechannel-discovery-e2e

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-extldflags -static -s" -o build/upgradechannel-discovery

.PHONY: build-docker
build-docker:
	DOCKER_BUILDKIT=1 docker build \
		-t ${REPO}:${TAG} .

.PHONY: build-docker-push
build-docker-push: build-docker
	docker push ${REPO}:${TAG}

.PHONY: test_deps
test_deps:
	go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
	go install github.com/onsi/gomega/...

.PHONY: unit-tests
unit-tests: test_deps
	ginkgo -r -v  --covermode=atomic --coverprofile=coverage.out -p -r ./pkg/...


e2e-tests:
	KUBE_VERSION=${KUBE_VERSION} $(ROOT_DIR)/scripts/e2e-tests.sh

kind-create-cluster:
	KUBE_VERSION=${KUBE_VERSION} $(ROOT_DIR)/scripts/create-cluster.sh

kind-e2e-tests: build-docker
	kind load docker-image --name $(CLUSTER_NAME) ${REPO}:${TAG}
	$(MAKE) e2e-tests