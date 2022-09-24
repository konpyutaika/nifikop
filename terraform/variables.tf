# NiFiKop
variable "nifikop_name" {
  description = "nifikop instance name"
  type        = string
  default     = "nifikop"
}

variable "nifikop_namespace" {
  description = "nifikop's namespace"
  type        = string
}

variable "nifikop_image_tag" {
  description = "nifikop's image tag"
  type        = string
  default     = "v0.14.0-release"
}


variable "nifikop_create_namespace" {
  description = "Whether or not we create the nifikop's namespace"
  type        = bool
  default     = true
}

variable "nifikop_watch_namespaces_list" {
  description = "list of namespaces watched by NiFiKop"
  type        = list(string)
}

variable "nifikop_config" {
  description = "nifikop chart helm configuration"
  type        = map(string)
}

variable "nifikop_version" {
  description = "Version nifikop"
  type        = string
  default     = "0.14.0"
}


# Zookeeper
variable "zookeeper_namespace" {
  description = "zookeeper's namespace"
  type        = string
}

variable "zookeeper_create_namespace" {
  description = "Whether or not we create the zookeeper's namespace"
  type        = string
  default     = false
}

variable "zookeeper_config" {
  description = "zookeeper chart helm configuration"
  type        = map(string)
}

## cert-manager configurations
variable "cert_manager_namespace" {
  description = "cert-manager's namespace"
  type        = string
  default     = "cert-manager"
}

variable "cert_manager_config" {
  description = "cert-manager chart helm configuration"
  type        = map(string)
}

variable "cert_manager_create_namespace" {
  description = "Whether or not we create the cert-manager's namespace"
  type        = bool
  default     = true
}


variable "cert_manager_version" {
  description = "Cert Manager version"
  type        = string
  default     = "v1.7.2"
}

# NiFi registry
variable "enable_nifi_registry" {
  description = "Whether or not to deploy Nifi Registry ressources"
  type        = bool
  default     = true
}

variable "nifi_registry_backend" {
  description = "Nifi registry backend type"
  type        = string
  default     = "db"

  validation {
    condition     = contains(["git", "db"], var.nifi_registry_backend)
    error_message = "Allowed values for nifi_registry_backend are \"db\" or \"git\"."
  }
}

variable "nifi_registry_image" {
  description = "nifi registry docker image to use"
  type        = string
}

variable "nifi_registry_sidecars" {
  description = "nifi registry sidecars config"
  type        = list(any)
  default     = []
}

variable "nifi_registry_namespace" {
  description = "nifi registry's namespace"
  type        = string
}

variable "nifi_registry_sa_annotations" {
  description = "A map of string to add as annotations."
  type        = map(string)
  default     = {}
}

variable "nifi_registry_node_selector_node_pool" {
  description = "nifi registry's node pool to use to deploy pod"
  type        = string
  default     = ""
}

variable "nifi_registry_container_port" {
  description = "nifi registry's namespace"
  type        = number
  default     = 18080
}

variable "nifi_registry_database_config" {
  description = "Configuration of nifi registry for database backend"
  type = object({
    url          = string,
    driver_class = string,
    user         = string,
    password     = string
  })
  default = {
    url          = ""
    driver_class = ""
    user         = ""
    password     = ""
  }
}

variable "nifi_registry_database_ssl_config" {
  description = "Configuration of nifi registry for ssl with database"
  type = object({
    cert           = string,
    private_key    = string,
    server_ca_cert = string,
  })
  default = {
    cert           = ""
    private_key    = ""
    server_ca_cert = ""
  }
}

variable "nifi_registry_git_config" {
  description = "Configuration of nifi registry for git backend"
  type = object({
    username             = string,
    user_email           = string,
    remote_url           = string,
    remote_branch        = string,
    remote_to_push       = string,
    ssh_known_hosts_path = string,
    ssh_key_path         = string
  })

  default = {
    username             = ""
    user_email           = ""
    remote_url           = ""
    remote_branch        = "master"
    remote_to_push       = "origin"
    ssh_known_hosts_path = ""
    ssh_key_path         = ""

  }
}

variable "nifi_registry_svc_annotations" {
  description = "Map of string(string) containing a set of annotations to add to the nifi registry's service"
  type        = map(string)
  default = {
    "cloud.google.com/load-balancer-type" = "Internal"
  }
}

variable "nifi_registry_service_type" {
  description = "Service type for the nifi registry"
  type        = string
  default     = "ClusterIP"

  validation {
    condition     = contains(["ClusterIP", "LoadBalancer", "NodePort"], var.nifi_registry_service_type)
    error_message = "Allowed values for nifi_registry_service_type are \"ClusterIP\", \"LoadBalancer\", or \"NodePort\"."
  }
}


# Helm

variable "force_update" {
  default     = false
  type        = bool
  description = "If true will force update an helm_release"
}
