---
id: 3_with_olm
title: With OLM
sidebar_label: With OLM
---

OLM is used to manage lifecycle of the Operator, and is also used to puclish on https://operatorhub.io

## Create new OLM release

You can create new version of CassKop OLM bundle using:

Exemple for generating version 0.0.4

```sh
operator-sdk olm-catalog gen-csv --csv-version 0.4.0 --update-crds
```

> You may need to manually update some fileds (such as description..), you can refere to previous versions for that

## Instruction to tests locally with OLM

Before submitting the operator to operatorhub.io you need to install and test OLM on a local Kubernetes.

These tests and all pre-requisite can also be executed automatically in a single step using a
[Makefile](https://github.com/operator-framework/community-operators/blob/master/docs/using-scripts.md).

Go to github/operator-framework/community-operators to interract with the OLM makefile

Install OLM

```sh
make operator.olm.install
```

Launch lint

```sh
make operator.verify OP_PATH=community-operators/casskop VERBOSE=true
```

Launch tests

```sh
make operator.test OP_PATH=community-operators/casskop VERBOSE=true
``` 