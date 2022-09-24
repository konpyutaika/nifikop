<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 0.13 |
| <a name="requirement_helm"></a> [helm](#requirement\_helm) | >= 2.5.1 |
| <a name="requirement_k8s"></a> [k8s](#requirement\_k8s) | 0.9.1 |
| <a name="requirement_kubernetes"></a> [kubernetes](#requirement\_kubernetes) | >= 2.0.2 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_helm"></a> [helm](#provider\_helm) | >= 2.5.1 |
| <a name="provider_k8s"></a> [k8s](#provider\_k8s) | 0.9.1 |
| <a name="provider_kubernetes"></a> [kubernetes](#provider\_kubernetes) | >= 2.0.2 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [helm_release.cert_manager](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release) | resource |
| [helm_release.nifikop](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release) | resource |
| [helm_release.zookeeper](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release) | resource |
| [k8s_manifest.cert_manager_crds](https://registry.terraform.io/providers/banzaicloud/k8s/0.9.1/docs/resources/manifest) | resource |
| [k8s_manifest.nifi_registry_deployment](https://registry.terraform.io/providers/banzaicloud/k8s/0.9.1/docs/resources/manifest) | resource |
| [k8s_manifest.nifi_registry_svc](https://registry.terraform.io/providers/banzaicloud/k8s/0.9.1/docs/resources/manifest) | resource |
| [k8s_manifest.nifikop_crds](https://registry.terraform.io/providers/banzaicloud/k8s/0.9.1/docs/resources/manifest) | resource |
| [kubernetes_cluster_role.zookeeper](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/cluster_role) | resource |
| [kubernetes_cluster_role_binding.zookeeper](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/cluster_role_binding) | resource |
| [kubernetes_namespace.cert_manager](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/namespace) | resource |
| [kubernetes_namespace.nifikop](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/namespace) | resource |
| [kubernetes_namespace.zookeeper](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/namespace) | resource |
| [kubernetes_network_policy.zookeeper](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/network_policy) | resource |
| [kubernetes_pod_security_policy.zookeeper](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/pod_security_policy) | resource |
| [kubernetes_secret.nifi_registry_secret](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/secret) | resource |
| [kubernetes_service_account.nifi_registry](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_account) | resource |
| [kubernetes_service_account.zookeeper](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/service_account) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_cert_manager_config"></a> [cert\_manager\_config](#input\_cert\_manager\_config) | cert-manager chart helm configuration | `map(string)` | n/a | yes |
| <a name="input_cert_manager_create_namespace"></a> [cert\_manager\_create\_namespace](#input\_cert\_manager\_create\_namespace) | Whether or not we create the cert-manager's namespace | `bool` | `true` | no |
| <a name="input_cert_manager_namespace"></a> [cert\_manager\_namespace](#input\_cert\_manager\_namespace) | cert-manager's namespace | `string` | `"cert-manager"` | no |
| <a name="input_cert_manager_version"></a> [cert\_manager\_version](#input\_cert\_manager\_version) | Cert Manager version | `string` | `"v1.7.2"` | no |
| <a name="input_enable_nifi_registry"></a> [enable\_nifi\_registry](#input\_enable\_nifi\_registry) | Whether or not to deploy Nifi Registry ressources | `bool` | `true` | no |
| <a name="input_force_update"></a> [force\_update](#input\_force\_update) | If true will force update an helm\_release | `bool` | `false` | no |
| <a name="input_nifi_registry_backend"></a> [nifi\_registry\_backend](#input\_nifi\_registry\_backend) | Nifi registry backend type | `string` | `"db"` | no |
| <a name="input_nifi_registry_container_port"></a> [nifi\_registry\_container\_port](#input\_nifi\_registry\_container\_port) | nifi registry's namespace | `number` | `18080` | no |
| <a name="input_nifi_registry_database_config"></a> [nifi\_registry\_database\_config](#input\_nifi\_registry\_database\_config) | Configuration of nifi registry for database backend | <pre>object({<br>    url          = string,<br>    driver_class = string,<br>    user         = string,<br>    password     = string<br>  })</pre> | <pre>{<br>  "driver_class": "",<br>  "password": "",<br>  "url": "",<br>  "user": ""<br>}</pre> | no |
| <a name="input_nifi_registry_database_ssl_config"></a> [nifi\_registry\_database\_ssl\_config](#input\_nifi\_registry\_database\_ssl\_config) | Configuration of nifi registry for ssl with database | <pre>object({<br>    cert           = string,<br>    private_key    = string,<br>    server_ca_cert = string,<br>  })</pre> | <pre>{<br>  "cert": "",<br>  "private_key": "",<br>  "server_ca_cert": ""<br>}</pre> | no |
| <a name="input_nifi_registry_git_config"></a> [nifi\_registry\_git\_config](#input\_nifi\_registry\_git\_config) | Configuration of nifi registry for git backend | <pre>object({<br>    username             = string,<br>    user_email           = string,<br>    remote_url           = string,<br>    remote_branch        = string,<br>    remote_to_push       = string,<br>    ssh_known_hosts_path = string,<br>    ssh_key_path         = string<br>  })</pre> | <pre>{<br>  "remote_branch": "master",<br>  "remote_to_push": "origin",<br>  "remote_url": "",<br>  "ssh_key_path": "",<br>  "ssh_known_hosts_path": "",<br>  "user_email": "",<br>  "username": ""<br>}</pre> | no |
| <a name="input_nifi_registry_image"></a> [nifi\_registry\_image](#input\_nifi\_registry\_image) | nifi registry docker image to use | `string` | n/a | yes |
| <a name="input_nifi_registry_namespace"></a> [nifi\_registry\_namespace](#input\_nifi\_registry\_namespace) | nifi registry's namespace | `string` | n/a | yes |
| <a name="input_nifi_registry_node_selector_node_pool"></a> [nifi\_registry\_node\_selector\_node\_pool](#input\_nifi\_registry\_node\_selector\_node\_pool) | nifi registry's node pool to use to deploy pod | `string` | `""` | no |
| <a name="input_nifi_registry_sa_annotations"></a> [nifi\_registry\_sa\_annotations](#input\_nifi\_registry\_sa\_annotations) | A map of string to add as annotations. | `map(string)` | `{}` | no |
| <a name="input_nifi_registry_service_type"></a> [nifi\_registry\_service\_type](#input\_nifi\_registry\_service\_type) | Service type for the nifi registry | `string` | `"ClusterIP"` | no |
| <a name="input_nifi_registry_sidecars"></a> [nifi\_registry\_sidecars](#input\_nifi\_registry\_sidecars) | nifi registry sidecars config | `list(any)` | `[]` | no |
| <a name="input_nifi_registry_svc_annotations"></a> [nifi\_registry\_svc\_annotations](#input\_nifi\_registry\_svc\_annotations) | Map of string(string) containing a set of annotations to add to the nifi registry's service | `map(string)` | <pre>{<br>  "cloud.google.com/load-balancer-type": "Internal"<br>}</pre> | no |
| <a name="input_nifikop_config"></a> [nifikop\_config](#input\_nifikop\_config) | nifikop chart helm configuration | `map(string)` | n/a | yes |
| <a name="input_nifikop_create_namespace"></a> [nifikop\_create\_namespace](#input\_nifikop\_create\_namespace) | Whether or not we create the nifikop's namespace | `bool` | `true` | no |
| <a name="input_nifikop_image_tag"></a> [nifikop\_image\_tag](#input\_nifikop\_image\_tag) | nifikop's image tag | `string` | `"v0.14.0-release"` | no |
| <a name="input_nifikop_name"></a> [nifikop\_name](#input\_nifikop\_name) | nifikop instance name | `string` | `"nifikop"` | no |
| <a name="input_nifikop_namespace"></a> [nifikop\_namespace](#input\_nifikop\_namespace) | nifikop's namespace | `string` | n/a | yes |
| <a name="input_nifikop_version"></a> [nifikop\_version](#input\_nifikop\_version) | Version nifikop | `string` | `"0.14.0"` | no |
| <a name="input_nifikop_watch_namespaces_list"></a> [nifikop\_watch\_namespaces\_list](#input\_nifikop\_watch\_namespaces\_list) | list of namespaces watched by NiFiKop | `list(string)` | n/a | yes |
| <a name="input_zookeeper_config"></a> [zookeeper\_config](#input\_zookeeper\_config) | zookeeper chart helm configuration | `map(string)` | n/a | yes |
| <a name="input_zookeeper_create_namespace"></a> [zookeeper\_create\_namespace](#input\_zookeeper\_create\_namespace) | Whether or not we create the zookeeper's namespace | `string` | `false` | no |
| <a name="input_zookeeper_namespace"></a> [zookeeper\_namespace](#input\_zookeeper\_namespace) | zookeeper's namespace | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_nifi_registry_container_port"></a> [nifi\_registry\_container\_port](#output\_nifi\_registry\_container\_port) | NiFi registry container port |
| <a name="output_nifi_registry_name"></a> [nifi\_registry\_name](#output\_nifi\_registry\_name) | NiFi registry name |
| <a name="output_nifi_registry_namespace"></a> [nifi\_registry\_namespace](#output\_nifi\_registry\_namespace) | NiFi registry namespace |
| <a name="output_nifi_registry_service_port"></a> [nifi\_registry\_service\_port](#output\_nifi\_registry\_service\_port) | NiFi registry service port |
| <a name="output_nifikop_image_tag"></a> [nifikop\_image\_tag](#output\_nifikop\_image\_tag) | The version of the NiFiKop instance |
| <a name="output_nifikop_name"></a> [nifikop\_name](#output\_nifikop\_name) | NiFiKop names |
| <a name="output_nifikop_namespace"></a> [nifikop\_namespace](#output\_nifikop\_namespace) | NiFiKop namespace |
| <a name="output_zookeeper_name"></a> [zookeeper\_name](#output\_zookeeper\_name) | Zookeeper name |
| <a name="output_zookeeper_namespace"></a> [zookeeper\_namespace](#output\_zookeeper\_namespace) | Zookeeper namespace |
<!-- END_TF_DOCS -->