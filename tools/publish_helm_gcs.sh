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
readonly HELM_TARBALL=helm-v3.4.2-linux-amd64.tar.gz
#readonly HELM_TARBALL=helm-v2.9.1-darwin-amd64.tar.gz
#readonly STABLE_REPO_URL=https://orange-kubernetes-charts.storage.googleapis.com/
readonly INCUBATOR_REPO_URL=https://orange-kubernetes-charts-incubator.storage.googleapis.com/
#readonly GCS_BUCKET_STABLE=gs://orange-kubernetes-charts
readonly GCS_BUCKET_INCUBATOR=gs://orange-kubernetes-charts-incubator

main() {
    mkdir -p tmp
    setup_helm_client
    authenticate

#    if ! sync_repo stable "$GCS_BUCKET_STABLE" "$STABLE_REPO_URL"; then
#        log_error "Not all stable charts could be packaged and synced!"
#    fi
    if ! sync_repo ${HELM_TARGET_DIR} "$GCS_BUCKET_INCUBATOR" "$INCUBATOR_REPO_URL" "$CHART_VERSION"; then
        log_error "Not all incubator charts could be packaged and pushed!"
    fi
}

setup_helm_client() {
    echo "Setting up Helm client..."

    curl --user-agent curl-ci-sync -sSL -o "$HELM_TARBALL" "$HELM_URL/$HELM_TARBALL"
    tar xzfv "$HELM_TARBALL" -C tmp
    rm -f "$HELM_TARBALL"

#    PATH="$(pwd)/tmp/darwin-amd64/:$PATH"
    PATH="$(pwd)/tmp/linux-amd64/:$PATH"

#    helm init --client-only
    helm repo add orange-incubator "$INCUBATOR_REPO_URL"
}

authenticate() {
    echo "Authenticating with Google Cloud..."
    gcloud auth activate-service-account --key-file <(base64 --decode <<< "$GCP_SA_CREDS")
}

sync_repo() {
    local target_dir="${1?Specify repo dir}"
    local bucket="${2?Specify repo bucket}"
    local repo_url="${3?Specify repo url}"
    local chart_version="${4?Specify chart version}"
    local index_dir="${target_dir}-index"


    echo "Syncing repo '$target_dir'..."

    mkdir -p "$target_dir"
    if ! gsutil cp "$bucket/index.yaml" "$index_dir/index.yaml"; then
        log_error "Exiting because unable to copy index locally. Not safe to proceed."
        exit 1
    fi

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

    if helm repo index --url "$repo_url" --merge "$index_dir/index.yaml" "$target_dir"; then
        # Move updated index.yaml to index folder so we don't push the old one again
        mv -f "$target_dir/index.yaml" "$index_dir/index.yaml"

        gsutil -h "Cache-Control:no-cache,max-age=0" cp "$target_dir/*" "$bucket"

        # Make sure index.yaml is synced last
        gsutil  -h "Cache-Control:no-cache,max-age=0" cp "$index_dir/index.yaml" "$bucket"
    else
        log_error "Exiting because unable to update index. Not safe to push update."
        exit 1
    fi

    ls -l "$target_dir"

    return "$exit_code"
}

log_error() {
    printf '\e[31mERROR: %s\n\e[39m' "$1" >&2
}

main
