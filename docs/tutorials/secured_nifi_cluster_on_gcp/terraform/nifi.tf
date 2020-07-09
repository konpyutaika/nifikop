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

resource "null_resource" "deploy-nifikop-crds" {
  depends_on = [kubernetes_namespace.nifi]
  provisioner "local-exec" {
    command = "gcloud container clusters get-credentials ${google_container_cluster.nifi-cluster.name}  --zone ${var.zone} --project ${var.project} && kubectl apply -f ../kubernetes/nifikop"
  }
}

// helm release
resource "helm_release" "nifikop" {
  name             = "nifikop"
  repository       = "https://orange-kubernetes-charts-incubator.storage.googleapis.com"
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
}