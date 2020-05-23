resource "kubernetes_storage_class" "nifi-ssd" {
  reclaim_policy      = "Delete"
  storage_provisioner = "kubernetes.io/gce-pd"
  volume_binding_mode = "WaitForFirstConsumer"

  parameters = {
    type = "pd-ssd"
  }

  metadata {
    annotations      = {}
    labels           = {}
    name             = "ssd-wait"
  }
  depends_on = [google_container_node_pool.nodes]
}

// Define prod namespace, for tracking pods deployment in production
resource "kubernetes_namespace" "nifi" {
  metadata {
    annotations = {
      name = var.nifi_namespace
    }
    # Enable istio sidecar injection, into pods instantiate into this namespace.
    labels = {
      istio-injection = "enabled"
      istio-operator-managed-injection = "enabled"
    }
    name = var.nifi_namespace
  }
  depends_on = [google_container_node_pool.nodes]
}

// helm release
resource "helm_release" "nifikop" {
  name             = "nifikop"
  repository       = data.helm_repository.orange-incubator.metadata[0].name
  chart            = "nifikop"
  version          = var.nifikop_chart_version
  namespace        = kubernetes_namespace.nifi.metadata[0].name
  disable_webhooks = false

  # Image configuration
  set {
    name  = "image.repository"
    value = var.nifikop_image_repo
  }

  set {
    name  = "image.tag"
    value = var.nifikop_image_tag
  }

  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding, /*helm_release.istio-operator,*/ helm_release.cert-manager]
}