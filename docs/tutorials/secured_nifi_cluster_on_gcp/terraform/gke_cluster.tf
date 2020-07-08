###################################################
##               GKE configuration               ##
## - Default pool : no nodes                     ##
## - VPA          : disabled                     ##
###################################################
// Define GKE cluster
resource "google_container_cluster" "nifi-cluster" {
  provider = google-beta
  name = "nifi-cluster"
  location =var.zone
  description = "GKE cluster for secured NiFi cluster"

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count = 1

  master_auth {
    username = var.username
    password = var.password

    client_certificate_config {
      issue_client_certificate = false
    }
  }

  node_locations = []

  # The configuration for addons supported by GKE
  addons_config {
    horizontal_pod_autoscaling { disabled = false }
    http_load_balancing { disabled = false }
    network_policy_config { disabled = true }
    istio_config { disabled = true }
    cloudrun_config { disabled = true }
  }

  # If enabled, all container images will be validated by Google Binary Authorization
  enable_binary_authorization = null

  # Whether Intra-node visibility is enabled for this cluster. This makes same node pod to pod traffic visible for VPC network
  enable_intranode_visibility = null

  # Whether to enable Kubernetes Alpha features for this cluster
  enable_kubernetes_alpha = false

  # Whether the ABAC authorizer is enabled for this cluster. When enabled, identities in the system, including service accounts, nodes, and controllers, will have statically granted permissions beyond those provided by the RBAC configuration or IAM
  enable_legacy_abac = false

  # Whether to enable Cloud TPU resources in this cluster
  enable_tpu = null

  # The logging service that the cluster should write logs to. Available options include logging.googleapis.com, logging.googleapis.com/kubernetes, and none
  logging_service = "logging.googleapis.com/kubernetes"

  # The monitoring service that the cluster should write metrics to
  monitoring_service = "monitoring.googleapis.com/kubernetes"

  # The minimum version of the master
  min_master_version = "1.16.9-gke.6"


  # The name or self_link of the Google Compute Engine network to which the cluster is connected. For Shared VPC, set this to the self link of the shared network
  network = "projects/${var.project}/global/networks/default"

  # Configuration options for the NetworkPolicy feature
  network_policy {
    enabled = false
    provider = "PROVIDER_UNSPECIFIED"
  }

  database_encryption {
    state = "DECRYPTED"
    key_name = ""
  }

  vertical_pod_autoscaling { enabled = false }
}

data "google_client_config" "current" {}

output "cluster_name" {
  value = "gke_${google_container_cluster.nifi-cluster.project}_${google_container_cluster.nifi-cluster.location}_${google_container_cluster.nifi-cluster.name}"
}

output "cluster_ca" {
  value = base64decode(google_container_cluster.nifi-cluster.master_auth.0.cluster_ca_certificate)
}

output "cluster_zone" {
  value = google_container_cluster.nifi-cluster.location
}

output "cluster_server" {
  value = "https://${google_container_cluster.nifi-cluster.endpoint}"
}