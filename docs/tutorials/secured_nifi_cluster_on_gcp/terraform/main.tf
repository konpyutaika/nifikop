terraform {
  required_providers {
    helm = "~> 0.10.4"
    kubernetes = "~> 1.10.0"
  }
}

// Provider definition
provider "google" {
  project = var.project
  credentials = file(var.service_account_json_file)
  region  = var.region
  zone    = var.zone
}

// Provider definition for beta features
provider "google-beta" {
  project = var.project
  credentials = file(var.service_account_json_file)
  region  = var.region
  zone    = var.zone
}

// Define Helm provider
provider "helm" {
  install_tiller  = true
  tiller_image    = "gcr.io/kubernetes-helm/tiller:${var.helm_version}"
  service_account = kubernetes_service_account.tiller.metadata.0.name
  debug           = true

  kubernetes {
    host                   = google_container_cluster.source-squidflow-cluster.endpoint
    token                  = data.google_client_config.current.access_token
    client_certificate     = base64decode(google_container_cluster.source-squidflow-cluster.master_auth.0.client_certificate)
    client_key             = base64decode(google_container_cluster.source-squidflow-cluster.master_auth.0.client_key)
    cluster_ca_certificate = base64decode(google_container_cluster.source-squidflow-cluster.master_auth.0.cluster_ca_certificate)
  }
}
