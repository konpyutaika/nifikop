---
id: 3_kubectl_plugin
title: Kubectl Plugin
sidebar_label: Kubectl Plugin
---

You can install the plugin by copying the [file](https://raw.githubusercontent.com/konpyutaika/nifikop/master/plugins/kubectl-nifikop) into your PATH.

For example on a UNIX machine:

```console
sudo cp plugins/kubectl-nifikop /usr/local/bin/kubectl-nifikop && sudo chmod +x /usr/local/bin/kubectl-nifikop
```

Then you can test the plugin:

```console
$ kubectl nifikop
usage: kubectl-nifikop <command> [<args>]
The available commands are:
   stop
   unstop
   start
   unstart
   stop_io
   unstop_io
For more information you can run kubectl-nifikop <command> --help
kubectl-nifikop: error: the following arguments are required: command
```

Your NiFiKop plugin is now installed!