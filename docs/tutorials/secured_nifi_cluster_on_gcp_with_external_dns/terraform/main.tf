data "google_client_config" "provider" {}

provider "kubernetes" {
  load_config_file = false

  host  = "https://${google_container_cluster.nifi-cluster.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(google_container_cluster.nifi-cluster.master_auth[0].cluster_ca_certificate, )
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

