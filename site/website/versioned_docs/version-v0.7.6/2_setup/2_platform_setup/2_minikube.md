---
id: 2_minikube
title: MiniKube
sidebar_label: MiniKube
---

Follow these instructions to prepare minikube for NiFiKop installation with sufficient resources to run NiFiKop and some basic applications.

## Prerequisites

- Administrative privileges are required to run minikube.

## Installation steps

1. Install the latest version of [minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/), version 1.1.1 or later, and a [minikube hypervisor driver](https://kubernetes.io/docs/tasks/tools/install-minikube/#install-a-hypervisor).
2. If you’re not using the default driver, set your minikube hypervisor driver.
   For example, if you installed the KVM hypervisor, set the vm-driver within the minikube configuration using the following command:

   ```sh
   minikube config set vm-driver kvm2
   ```

3. Start minikube with 16384 MB of memory and 4 CPUs. This example uses Kubernetes version 1.14.2. You can change the version to any Kubernetes version supported by NiFiKop by altering the --kubernetes-version value:

   ```sh
   $ minikube start --memory=16384 --cpus=4 --kubernetes-version=v1.14.2
   ```

Depending on the hypervisor you use and the platform on which the hypervisor is run, minimum memory requirements vary. 16384 MB is sufficent to run NiFiKop.

:::tip
If you don’t have enough RAM allocated to the minikube virtual machine, the following errors could occur:

- Image pull failures
- Healthcheck timeout failures
- Kubectl failures on the host
- General network instability of the virtual machine and the host
- Complete lock-up of the virtual machine
- Host NMI watchdog reboots
- One effective way to monitor memory usage in minikube:

```sh
minikube ssh
top
```

:::
