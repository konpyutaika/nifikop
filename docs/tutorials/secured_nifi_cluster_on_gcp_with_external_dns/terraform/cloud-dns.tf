// Documentations support : https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/gke.md
// Manage DNS Zone for external dns.
resource "google_dns_managed_zone" "external-dns-zone" {
    count = var.create_dns ? 1 : 0

    name         = var.dns_zone_name
    dns_name     = "${var.dns_name}."
    description  = "Automatically managed zone by kubernetes.io/external-dns"
    labels       = {}
    visibility   = "public"
}

// Manage DNS record sets for external dns zone.
resource "google_dns_record_set" "external-dns-record-set" {
    count = var.create_dns ? 1 : 0
    managed_zone = var.managed_zone
    name         = google_dns_managed_zone.external-dns-zone[0].dns_name
    rrdatas      = google_dns_managed_zone.external-dns-zone[0].name_servers
    ttl          = 300
    type         = "NS"
}


// Cluster role with permissions required for external dns.
resource "kubernetes_cluster_role" "external-dns" {
  metadata {
    name = "external-dns"
  }

  rule {
    api_groups = ["",]
    resources  = ["services",]
    verbs       = ["get", "watch", "list",]
  }

  rule {
    api_groups = ["",]
    resources  = ["pods",]
    verbs       = ["get", "watch", "list",]
  }

  rule {
    api_groups = ["",]
    resources  = ["nodes",]
    verbs       = ["get", "watch", "list",]
  }

  rule {
    api_groups = ["extensions",]
    resources  = ["ingresses",]
    verbs      = ["get", "watch", "list",]
  }
  depends_on = [google_container_node_pool.nodes]
}

// Create service account for external-dns
resource "kubernetes_service_account" "external-dns" {
  metadata {
    name = "external-dns"
    namespace = kubernetes_namespace.nifi.metadata.0.name
  }
  automount_service_account_token = true
  depends_on = [google_container_node_pool.nodes]
}

// Binding external-dns cluster role, with the external-dns Service account.
resource "kubernetes_cluster_role_binding" "external-dns-viewer" {
  metadata {
    name      = "external-dns-viewer"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.external-dns.metadata.0.name
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.external-dns.metadata.0.name
    namespace = kubernetes_service_account.external-dns.metadata.0.namespace
  }
}

// Create deployment 
resource "kubernetes_deployment" "external-dns" {
  metadata {
    annotations      = {}
    labels           = {
      "app" = "external-dns"
    }
    name             = "external-dns"
    namespace = kubernetes_service_account.external-dns.metadata.0.namespace
  }

  spec {
    selector {
      match_labels = {
        "app" = "external-dns"
      }
    }

    strategy {
      type = "Recreate"
    }

    template {
      metadata {
        annotations = {}
        labels      = {
          "app" = "external-dns"
        }
      }

      spec {
        service_account_name = kubernetes_service_account.external-dns.metadata.0.name
        automount_service_account_token = true
        container {
          args = [
            "--source=service",
            "--source=ingress",
            "--domain-filter=${var.dns_name}",
            "--provider=google",
            "--policy=upsert-only",
            "--registry=txt",
            "--namespace=${kubernetes_namespace.nifi.metadata[0].name}",
            "--fqdn-template={{.Name}}.${var.dns_name}",
            "--txt-owner-id=gke-${google_container_cluster.nifi-cluster.project}-${google_container_cluster.nifi-cluster.location}_${google_container_cluster.nifi-cluster.name}-${kubernetes_namespace.nifi.metadata[0].name}",
          ]
          command                  = []
          image                    = "bitnami/external-dns:0.5.14"
          name                     = "external-dns"
        }
      }
    }
  }
}
