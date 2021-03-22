# Image URL to use all building/pushing image targets
SERVICE_NAME			:= nifikop
DOCKER_REGISTRY_BASE 	?= orangeopensource
IMAGE_TAG				?= $(shell git describe --tags --abbrev=0 --match '[0-9].*[0-9].*[0-9]' 2>/dev/null)
IMAGE_NAME 				?= $(SERVICE_NAME)
BUILD_IMAGE				?= orangeopensource/nifikop-build
GOLANG_VERSION          ?= 1.15

# workdir
WORKDIR := /go/nifikop

# Debug variables
TELEPRESENCE_REGISTRY ?= datawire

DEV_DIR := docker/build-image

# Repository url for this project
# in gitlab CI_REGISTRY_IMAGE=repo/path/name:tag
ifdef CI_REGISTRY_IMAGE
	REPOSITORY := $(CI_REGISTRY_IMAGE)
else
	REPOSITORY := $(DOCKER_REGISTRY_BASE)/$(IMAGE_NAME)
endif

# Branch is used for the docker image version
ifdef CIRCLE_BRANCH
#	removing / for fork which lead to docker error
	BRANCH := $(subst /,-,$(CIRCLE_BRANCH))
else
	ifdef CIRCLE_TAG
		BRANCH := $(CIRCLE_TAG)
	else
		BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
	endif
endif

# Operator version is managed in go file
# BaseVersion is for dev docker image tag
BASEVERSION := $(shell awk -F\" '/Version =/ { print $$2}' version/version.go)

ifdef CIRCLE_TAG
	VERSION := ${BRANCH}
else
	VERSION := $(BASEVERSION)-${BRANCH}
endif

HELM_VERSION    := $(shell cat helm/nifikop/Chart.yaml| grep version | awk -F"version: " '{print $$2}')
HELM_TARGET_DIR ?= docs/helm

# if branch master tag latest
ifeq ($(CIRCLE_BRANCH),master)
	PUSHLATEST := true
endif

# The default action of this Makefile is to build the development docker image
default: build
all: manager

# Default bundle image tag
BUNDLE_IMG ?= $(REPOSITORY)-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)


# ----
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Generate code
generate: controller-gen
	@echo "Generate zzz-deepcopy objects"
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Build the docker image
docker-build:
	docker build -t $(REPOSITORY):$(VERSION) .

build: manager manifests docker-build

# Download controller-gen locally if necessary
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out; go tool cover -html=cover.out -o coverage.html

test-with-vendor: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test -mod=vendor ./... -coverprofile cover.out; go tool cover -html=cover.out -o coverage.html

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(REPOSITORY):$(VERSION)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Download kustomize locally if necessary
KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(REPOSITORY):$(VERSION)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

helm-package:
	@echo Packaging $(CHART_VERSION)
ifdef CHART_VERSION
	    echo $(CHART_VERSION)
	    helm package --version $(CHART_VERSION) helm/nifikop
else
		CHART_VERSION=$(HELM_VERSION)
	    helm package helm/nifikop
endif
	mv nifikop-$(CHART_VERSION).tgz $(HELM_TARGET_DIR)
	helm repo index $(HELM_TARGET_DIR)/

# Push the docker image
docker-push:
	docker push $(REPOSITORY):$(VERSION)
ifdef PUSHLATEST
	docker tag $(REPOSITORY):$(VERSION) $(REPOSITORY):latest
	docker push $(REPOSITORY):latest
endif
# ----

.DEFAULT_GOAL := help
help:
	@grep -E '(^[a-zA-Z_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

get-version:
	@echo $(VERSION)

clean:
	@rm -rf $(OUT_BIN) || true
	@rm -f api/v1alpha1/zz_generated.deepcopy.go || true

#Generate dep for graph
UNAME := $(shell uname -s)

dep-graph:
ifeq ($(UNAME), Darwin)
	dep status -dot | dot -T png | open -f -a /Applications/Preview.app
endif
ifeq ($(UNAME), Linux)
	dep status -dot | dot -T png | display
endif

debug-port-forward:
	kubectl port-forward `kubectl get pod -l app=nifikop -o jsonpath="{.items[0].metadata.name}"` 40000:40000

debug-pod-logs:
	kubectl logs -f `kubectl get pod -l app=nifikop -o jsonpath="{.items[0].metadata.name}"`

define debug_telepresence
	export TELEPRESENCE_REGISTRY=$(TELEPRESENCE_REGISTRY) ; \
	echo "execute : cat nifi-operator.env" ; \
	sudo mkdir -p /var/run/secrets/kubernetes.io ; \
	tdep=$(shell kubectl get deployment -l app=nifikop -o jsonpath='{.items[0].metadata.name}') ; \
  	echo kubectl get deployment -l app=nifikop -o jsonpath='{.items[0].metadata.name}' ; \
	echo telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file nifi-operator.env $1 $2 ; \
 	telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file nifi-operator.env $1 $2
endef

debug-telepresence:
	$(call debug_telepresence)

debug-telepresence-with-alias:
	$(call debug_telepresence,--also-proxy,10.40.0.0/16)

# Build the docker development environment
build-ci-image:
	docker build --cache-from $(BUILD_IMAGE):latest \
	  --build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
		-t $(BUILD_IMAGE):latest \
		-t $(BUILD_IMAGE):$(GOLANG_VERSION) \
		-f $(DEV_DIR)/Dockerfile \
		.

push-ci-image:
	docker push $(BUILD_IMAGE):$(GOLANG_VERSION)
ifdef PUSHLATEST
	docker push $(BUILD_IMAGE):latest
endif

## Test if the dependencies we need to run this Makefile are installed
#deps-development:
#ifndef DOCKER
#	@echo "Docker is not available. Please install docker"
#	@exit 1
#endif