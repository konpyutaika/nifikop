---
id: 3_kubectl_plugin
title: Kubectl Plugin
sidebar_label: Kubectl Plugin
---

You can build the plugin and copy the exectuable into your PATH.

For example on a UNIX machine:

```console
make kubectl-nifikop && sudo cp ./bin/kubectl-nifikop /usr/local/bin
```

Then you can test the plugin:

```console
$ kubectl nifikop
Usage:
  nifikop [command]

Available Commands:
  completion          Generate the autocompletion script for the specified shell
  help                Help about any command
  nificluster         
  nificonnection      
  nifidataflow        
  nifigroupautoscaler 
  nifiregistryclient  
  nifiuser            
  nifiusergroup 
```

Your NiFiKop plugin is now installed!