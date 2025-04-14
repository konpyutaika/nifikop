# Image URL to use all building/pushing image targets
SERVICE_NAME					:= nifikop
DOCKER_REGISTRY_BASE 	?= ghcr.io/konpyutaika/docker-images
IMAGE_TAG							?= $(shell git describe --tags --abbrev=0 --match '[0-9].*[0-9].*[0-9]' 2>/dev/null)
IMAGE_NAME 						?= $(SERVICE_NAME)
BUILD_IMAGE						?= ghcr.io/konpyutaika/docker-images/nifikop-build
GOLANG_VERSION        ?= 1.24.2
IMAGE_TAG_BASE 				?= <registry>/<operator name>
OS 										 = $(shell go env GOOS)
ARCH 									 = $(shell go env GOARCH)

# workdir
WORKDIR := /go/nifikop

# Debug variables
TELEPRESENCE_REGISTRY ?= datawire

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v5.4.3
CONTROLLER_TOOLS_VERSION ?= v0.16.1
ENVTEST_K8S_VERSION = 1.31.0
ENVTEST_VERSION ?= release-0.19
GOLANGCI_LINT_VERSION ?= v1.59.1

# controls the exit code of linter in case of linting issues
GOLANGCI_EXIT_CODE = 0

DEV_DIR := docker/build-image

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
	BUNDLE_GEN_FLAGS += --use-image-digests
endif


# Repository url for this project
# in gitlab CI_REGISTRY_IMAGE=repo/path/name:tag
ifdef CI_REGISTRY_IMAGE
	REPOSITORY := $(CI_REGISTRY_IMAGE)
else
	REPOSITORY := $(DOCKER_REGISTRY_BASE)/$(IMAGE_NAME)
endif
IMAGE_TAG_BASE = $(REPOSITORY)

# Branch is used for the docker image version
ifdef CIRCLE_BRANCH
#	removing / for fork which lead to docker error
	BRANCH := $(subst /,-,$(CIRCLE_BRANCH))
else
	ifdef CIRCLE_TAG
		BRANCH := $(CIRCLE_TAG)
	else
		BRANCH=$(shell git rev-parse --abbrev-ref HEAD | sed -e 's/\//_/')
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

HELM_VERSION     := $(shell cat helm/nifikop/Chart.yaml| grep version | awk -F"version: " '{print $$2}')
HELM_TARGET_DIR  ?= docs/helm
HELM_CHARTS_DIRS := $(wildcard helm/*/.)

# if branch master tag latest
ifeq ($(CIRCLE_BRANCH),master)
	PUSHLATEST := true
endif

# The default action of this Makefile is to build the development docker image
.PHONY: default
default: build

.PHONY: all
all: build

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

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Build manager binary
.PHONY: manager
manager: manifests generate fmt
	go build -o bin/manager cmd/main.go

# Generate code
.PHONY: generate
generate: controller-gen
	@echo "Generate zzz-deepcopy objects"
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd:maxDescLen=0 webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	mkdir -p helm/nifikop/crds && cp config/crd/bases/* helm/nifikop/crds

# Build the docker image
.PHONY: docker-build
docker-build:
	${CONTAINER_TOOL} build -t $(REPOSITORY):$(VERSION) .

.PHONY: build
build: manager manifests

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: lint
lint: golangci-lint
	 $(GOLANGCI_LINT) run --issues-exit-code $(GOLANGCI_EXIT_CODE)

# Run go fmt against code
.PHONY: fmt
fmt:
	@go mod tidy
	@go fmt ./...

# RUN https://go.dev/blog/vuln against code for known CVEs
.PHONY: govuln
govuln:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
.PHONY: test
test: manifests generate fmt lint helm-chart-version-match govuln envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: test-with-vendor
test-with-vendor: manifests generate fmt lint helm-chart-version-match govuln envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test -mod=vendor ./... -coverprofile cover.out

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: generate fmt manifests
	go run ./cmd/main.go

ifndef ignore-not-found
  ignore-not-found = false
endif

# Install CRDs into a cluster
.PHONY: install
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | ${KUBECTL} apply -f -

# Uninstall CRDs from a cluster
.PHONY: uninstall
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | ${KUBECTL} delete --ignore-not-found=$(ignore-not-found) -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(REPOSITORY):$(VERSION)
	$(KUSTOMIZE) build config/default | ${KUBECTL} apply -f -

# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
.PHONY: undeploy
undeploy:
	$(KUSTOMIZE) build config/default | ${KUBECTL} delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))


# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
	operator-sdk generate kustomize manifests --interactive=false -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(REPOSITORY):$(VERSION)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle $(BUNDLE_GEN_FLAGS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	${CONTAINER_TOOL} build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: helm-dep-update $(HELM_CHARTS_DIRS)
helm-dep-update: $(HELM_CHARTS_DIRS)
$(HELM_CHARTS_DIRS):
	@echo Updating helm chart $@ dependencies
	@cd $@ && helm dependency update

.PHONY: helm-gen-docs
helm-gen-docs:
# only generate docs for nifi-cluster to avoid stomping on the existing nifikop chart docs
	${CONTAINER_TOOL} run --rm --volume "$(shell pwd)/helm/nifi-cluster/:/helm-docs" -u $(shell id -u) jnorwood/helm-docs:latest

.PHONY: helm-chart-version-match
helm-chart-version-match:
	@echo Checking Chart.yaml\'s version values match
	@OPERATOR_VERSION=$$(grep '^version:' helm/nifikop/Chart.yaml | cut -d ' ' -f 2); \
	for dir in helm/*; do \
		if [ -f "$$dir/Chart.yaml" ]; then \
			CURRENT_VERSION=$$(grep '^version:' $$dir/Chart.yaml | cut -d ' ' -f 2); \
			if [ "$$OPERATOR_VERSION" != "$$CURRENT_VERSION" ]; then \
				echo "Version mismatch in $$dir/Chart.yaml: $$CURRENT_VERSION"; \
				exit 1; \
			fi; \
		fi; \
	done; \
	echo "All versions match: $$OPERATOR_VERSION"

.PHONY: helm-package
helm-package:
# package operator chart
	@echo Packaging NiFiKop $(CHART_VERSION)
ifdef CHART_VERSION
	    echo $(CHART_VERSION)
	    helm package --version $(CHART_VERSION) helm/nifikop
			helm dependency update helm/nifi-cluster
	    helm package --version $(CHART_VERSION) helm/nifi-cluster
else
		CHART_VERSION=$(HELM_VERSION) helm package helm/nifikop
		helm dependency update helm/nifi-cluster
		CHART_VERSION=$(HELM_VERSION) helm package helm/nifi-cluster
endif
	mv nifikop-$(CHART_VERSION).tgz $(HELM_TARGET_DIR)
	mv nifi-cluster-$(CHART_VERSION).tgz $(HELM_TARGET_DIR)
	helm repo index $(HELM_TARGET_DIR)/


# Push the docker image
.PHONY: docker-push
docker-push:
	${CONTAINER_TOOL} push $(REPOSITORY):$(VERSION)
ifdef PUSHLATEST
	${CONTAINER_TOOL} tag $(REPOSITORY):$(VERSION) $(REPOSITORY):latest
	${CONTAINER_TOOL} push $(REPOSITORY):latest
endif
# ----

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> than the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
  # copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- ${CONTAINER_TOOL} buildx create --name project-v3-builder
	${CONTAINER_TOOL} buildx use project-v3-builder
ifdef PUSHLATEST
	- ${CONTAINER_TOOL} buildx build --push --platform=$(PLATFORMS) --tag $(REPOSITORY):$(VERSION) --tag $(REPOSITORY):latest -f Dockerfile.cross .
else
	- ${CONTAINER_TOOL} buildx build --push --platform=$(PLATFORMS) --tag $(REPOSITORY):$(VERSION) -f Dockerfile.cross .
endif
	- ${CONTAINER_TOOL} buildx rm project-v3-builder
	rm Dockerfile.cross

.DEFAULT_GOAL := help

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: get-version
get-version:
	@echo $(VERSION)

.PHONY: clean
clean:
	@rm -rf $(OUT_BIN) || true
	@rm -f api/v1alpha1/zz_generated.deepcopy.go || true

#Generate dep for graph
UNAME := $(shell uname -s)

.PHONY: dep-graph
dep-graph:
ifeq ($(UNAME), Darwin)
	dep status -dot | dot -T png | open -f -a /Applications/Preview.app
endif
ifeq ($(UNAME), Linux)
	dep status -dot | dot -T png | display
endif

.PHONY: debug-port-forward
debug-port-forward:
	${KUBECTL} port-forward `${KUBECTL} get pod -l app=nifikop -o jsonpath="{.items[0].metadata.name}"` 40000:40000

.PHONY: debug-pod-logs
debug-pod-logs:
	${KUBECTL} logs -f `${KUBECTL} get pod -l app=nifikop -o jsonpath="{.items[0].metadata.name}"`

define debug_telepresence
	export TELEPRESENCE_REGISTRY=$(TELEPRESENCE_REGISTRY) ; \
	echo "execute: cat nifi-operator.env" ; \
	sudo mkdir -p /var/run/secrets/kubernetes.io ; \
	tdep=$(shell ${KUBECTL} get deployment -l app=nifikop -o jsonpath='{.items[0].metadata.name}') ; \
  	echo ${KUBECTL} get deployment -l app=nifikop -o jsonpath='{.items[0].metadata.name}' ; \
	echo telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file nifi-operator.env $1 $2 ; \
 	telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file nifi-operator.env $1 $2
endef

.PHONY: debug-telepresence
debug-telepresence:
	$(call debug_telepresence)

.PHONY: debug-telepresence-with-alias
debug-telepresence-with-alias:
	$(call debug_telepresence,--also-proxy,10.40.0.0/16)

# Build the docker development environment
.PHONY: build-ci-image
build-ci-image:
	${CONTAINER_TOOL} build --cache-from $(BUILD_IMAGE):latest \
	  --build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
	  --build-arg GOLANGCI_VERSION=$(GOLANGCI_LINT_VERSION) \
		-t $(BUILD_IMAGE):latest \
		-t $(BUILD_IMAGE):$(GOLANG_VERSION) \
		-f $(DEV_DIR)/Dockerfile \
		.

.PHONY: push-ci-image
push-ci-image:
	${CONTAINER_TOOL} push $(BUILD_IMAGE):$(GOLANG_VERSION)
ifdef PUSHLATEST
	${CONTAINER_TOOL} push $(BUILD_IMAGE):latest
endif

## Test if the dependencies we need to run this Makefile are installed
#deps-development:
#ifndef DOCKER
#	@echo "Docker is not available. Please install docker"
#	@exit 1
#endif

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.23.0/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

.PHONY: kubectl-nifikop
kubectl-nifikop:
	go build -o bin/kubectl-nifikop ./cmd/kubectl-nifikop/main.go

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

.PHONY: build-installer
build-installer: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > dist/install.yaml