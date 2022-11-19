#!/usr/bin/env bash

# Copyright 2018 The Kubernetes Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail


HELM_TARGET_DIR=$(pwd)/tmp/incubator
readonly HELM_URL=https://get.helm.sh
#readonly HELM_TARBALL=helm-v3.8.0-linux-amd64.tar.gz
readonly HELM_TARBALL=helm-v3.8.0-darwin-amd64.tar.gz
readonly INCUBATOR_REPO_URL=ghcr.io/konpyutaika

main() {
    mkdir -p tmp
    setup_helm_client
    authenticate

    if ! sync_repo ${HELM_TARGET_DIR} "oci://$INCUBATOR_REPO_URL/helm-charts" "$CHART_VERSION"; then
        log_error "Not all incubator charts could be packaged and pushed!"
    fi
}

setup_helm_client() {
    echo "Setting up Helm client..."

    curl --user-agent curl-ci-sync -sSL -o "$HELM_TARBALL" "$HELM_URL/$HELM_TARBALL"
    tar xzfv "$HELM_TARBALL" -C tmp
    rm -f "$HELM_TARBALL"

    PATH="$(pwd)/tmp/darwin-amd64/:$PATH"
#    PATH="$(pwd)/tmp/linux-amd64/:$PATH"
}

authenticate() {
    echo "Authenticating to Github packages ..."
    helm registry login -u $GH_NAME --password $GH_TOKEN $INCUBATOR_REPO_URL
}

sync_repo() {
    local target_dir="${1?Specify repo dir}"
    local repo_url="${2?Specify repo url}"
    local chart_version="${3?Specify chart version}"
    local index_dir="${target_dir}-index"


    echo "Syncing repo '$target_dir'..."

    mkdir -p "$target_dir"

    local exit_code=0

    echo "Packaging operators ..."
    if [[ -n "$chart_version" ]]; then
      if ! HELM_TARGET_DIR=${target_dir} make helm-package; then
        log_error "Problem packaging operator"
        exit_code=1
      fi
    else
      if ! CHART_VERSION=${chart_version} HELM_TARGET_DIR=${target_dir} make helm-package; then
        log_error "Problem packaging operator"
        exit_code=1
      fi
    fi

    if ! helm push ${target_dir}/nifikop-${CHART_VERSION}.tgz ${repo_url}; then
        log_error "Exiting because unable to push chart."
        exit 1
    fi

    ls -l "$target_dir"

    return "$exit_code"
}

log_error() {
    printf '\e[31mERROR: %s\n\e[39m' "$1" >&2
}

main
