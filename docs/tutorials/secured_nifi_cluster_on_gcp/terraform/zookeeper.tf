// Define prod namespace, for tracking pods deployment in production
resource "kubernetes_namespace" "zookeeper" {
  metadata {
    annotations = {
      name = var.zookeeper_namespace
    }
    # Enable istio sidecar injection, into pods instantiate into this namespace.
    labels = {
      istio-injection = "enabled"
      istio-operator-managed-injection = "enabled"
    }
    name = var.zookeeper_namespace
  }
  depends_on = [google_container_node_pool.nodes]
}

// helm release
resource "helm_release" "zookeeper" {
  name             = "zookeeper"
  repository       = data.helm_repository.bitnami.metadata[0].name
  chart            = "bitnami/zookeeper"
  namespace        = kubernetes_namespace.zookeeper.metadata[0].name
  disable_webhooks = false
  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding]
  set {
    name  = "replicaCount"
    value = 3
  }

  set {
    name  = "global.storageClass"
    value = kubernetes_storage_class.nifi-ssd.metadata[0].name
  }
}