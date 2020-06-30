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

resource "kubernetes_namespace" "nifi" {
  metadata {
    annotations = {
      name = var.nifi_namespace
    }
    name = var.nifi_namespace
  }
  depends_on = [google_container_node_pool.nodes]
}

// helm release
/*resource "helm_release" "nifikop" {
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

  set {
    name  = "certManager.namespace"
    value = "default"
  }


  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding]
}*/