#!/bin/bash

# only generate docs for nifi-cluster to avoid stomping on the existing nifikop chart docs
docker run --rm --volume "$(pwd)/nifi-cluster:/helm-docs" -u $(id -u) jnorwood/helm-docs:latest