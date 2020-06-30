resource "kubernetes_namespace" "zookeeper" {
  metadata {
    annotations = {
      name = var.zookeeper_namespace
    }
    name = var.zookeeper_namespace
  }
  depends_on = [google_container_node_pool.nodes]
}

// helm release
resource "helm_release" "zookeeper" {
  name             = "zookeeper"
  repository       = "https://charts.bitnami.com/bitnami"
  chart            = "zookeeper"
  namespace        = kubernetes_namespace.zookeeper.metadata[0].name
  disable_webhooks = false
  set {
    name  = "replicaCount"
    value = 3
  }

  set {
    name  = "global.storageClass"
    value = kubernetes_storage_class.nifi-ssd.metadata[0].name
  }
}