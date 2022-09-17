---
id: 1_developer_guide
title: Developer guide
sidebar_label: Developer guide
---

## Operator SDK

### Prerequisites

NiFiKop has been validated with :

- [go](https://golang.org/doc/install) version v1.17+.
- [docker](https://docs.docker.com/get-docker/) version 18.09+
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) version v1.16+
- [Helm](https://helm.sh/) version v3.4.2
- [Operator sdk](https://github.com/operator-framework/operator-sdk) version v1.18.1

### Initial setup

Checkout the project.

```bash
git clone https://github.com/konpyutaika/nifikop.git
cd nifikop
```

### Operator sdk

The full list of command is available here : https://sdk.operatorframework.io/docs/upgrading-sdk-version/v1.0.0/#cli-changes

### Build NiFiKop

#### Local environment

If you prefer working directly with your local go environment you can simply uses :

```bash
make build
```

### Run NiFiKop

We can quickly run NiFiKop in development mode (on your local host), then it will use your kubectl configuration file to connect to your kubernetes cluster.

There are several ways to execute your operator :

- Using your IDE directly
- Executing directly the Go binary
- deploying using the Helm charts

If you want to configure your development IDE, you need to give it environment variables so that it will uses to connect to kubernetes.

```bash
KUBECONFIG={path/to/your/kubeconfig}
WATCH_NAMESPACE={namespace_to_watch}
POD_NAME={name for operator pod}
LOG_LEVEL=Debug
OPERATOR_NAME=ide
```

#### Run the Operator Locally with the Go Binary

This method can be used to run the operator locally outside of the cluster. This method may be preferred during development as it facilitates faster deployment and testing.

Set the name of the operator in an environment variable

```bash
export OPERATOR_NAME=nifi-operator
```

Deploy the CRDs.

```bash
kubectl apply -f config/crd/bases/nifi.konpyutaika.com_nificlusters.yaml
kubectl apply -f config/crd/bases/nifi.konpyutaika.com_nifidataflows.yaml
kubectl apply -f config/crd/bases/nifi.konpyutaika.com_nifiparametercontexts.yaml
kubectl apply -f config/crd/bases/nifi.konpyutaika.com_nifiregistryclients.yaml
kubectl apply -f config/crd/bases/nifi.konpyutaika.com_nifiusergroups.yaml
kubectl apply -f config/crd/bases/nifi.konpyutaika.com_nifiusers.yaml
```

And deploy the operator.

```bash
make run
```

This will run the operator in the `default` namespace using the default Kubernetes config file at `$HOME/.kube/config`.

#### Deploy using the Helm Charts

This section provides an instructions for running the operator Helm charts with an image that is built from the local branch.

Build the image from the current branch.

```bash
export DOCKER_REPO_BASE={your-docker-repo}
make docker-build
```

Push the image to docker hub (or to whichever repo you want to use)

```bash
$ make docker-push
```

:::info
The image tag is a combination of the version as defined in `verion/version.go` and the branch name.
:::

Install the Helm chart.

```bash
helm install skeleton ./helm/nifikop \
    --set image.tag=v0.5.1-release \
    --namespace-{"nifikop"}
```

:::important
The `image.repository` and `image.tag` template variables have to match the names from the image that we pushed in the previous step.
:::

:::info
We set the chart name to the branch, but it can be anything.
:::

Lastly, verify that the operator is running.

```console
$ kubectl get pods -n nifikop
NAME                                                READY   STATUS    RESTARTS   AGE
skeleton-nifikop-8946b89dc-4cfs9   1/1     Running   0          7m45s
```

## Helm

The NiFiKop operator is released in the `konpyutaika-incubator` helm repository.

In order to package the chart you need to run the following command.

```bash
make helm-package
```
