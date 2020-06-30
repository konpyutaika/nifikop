// VARIABLES
variable "service_account_json_file" {
  description = "Path to local service account's json file"
  type        = string
}

# GCP configurations
variable "region" {
  description = "GCP region"
  type        = string
}

variable "zone" {
  description = "GCP zone"
  type        = string
}

variable "project" {
  description = "GCP project name"
  type        = string
}

# GKE variables
variable "username" {
  description = "GKE username"
  type        = string
}

variable "password" {
  description = "GKE password"
  type        = string
}

variable "min_node" {
  type        = number
  description = "Minimum number of nodes in the NodePool. Must be >=0 and <= max_node_count."
  default     = 1
}

variable "max_node" {
  type        = number
  description = "Maximum number of nodes in the NodePool. Must be >= min_node_count."
  default     = 1
}

variable "initial_node_count" {
  type        = number
  description = "The number of nodes to create in this cluster's default node pool."
  default     = 1
}

variable "preemptible" {
  type        = bool
  description = "true/false using preemptibles nodes."
}

variable "cluster_machines_types" {
  type        = string
  description = "Defines the machine type"
}

variable "nifi_namespace" {
  description = "Name of the namesapce associated to the nifi deployments"
  type        = string
  default     = "nifi"
}

# Cert-manager
variable "cert_manager_namespace" {
  description = "Cert-manager's namespace"
  type = string
  default = "cert-manager"
}

# Zookeeper
variable "zookeeper_namespace" {
  description = "Zookeeper's namespace"
  type = string
  default = "zookeeper"
}

# DNS

variable "create_dns" {
  description = "If set to 1, create cloud dns for external-DNS"
  type        = bool
}

variable "dns_zone_name" {
  description = ""
  type        = string
}

variable "dns_name" {
  description = ""
  type        = string
}

variable "managed_zone" {
  description = ""
  type        = string
}