---
id: 2_with_helm
title: With Helm
sidebar_label: With Helm
---

The CassKop operator is released in the helm/charts/incubator see : https://github.com/helm/charts/pull/14414

We also have a helm repository hosted on GitHub pages.

## Release helm charts on GitHub

In order to release the Helm charts on GitHub, we need to generate the package locally

```sh
make helm-package
```

then add to git the package and make a PR on the repo.
