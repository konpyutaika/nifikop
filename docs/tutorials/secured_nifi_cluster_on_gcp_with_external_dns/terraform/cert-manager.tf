resource "kubernetes_namespace" "cert-manager" {
  metadata {
    annotations = {
      name = var.cert_manager_namespace
    }
    name = var.cert_manager_namespace
  }
  depends_on = [google_container_node_pool.nodes]
}

resource "null_resource" "deploy-certmanager-crds" {
  depends_on = [kubernetes_namespace.cert-manager]
  provisioner "local-exec" {
    command = "gcloud container clusters get-credentials ${google_container_cluster.nifi-cluster.name}  --zone ${var.zone} --project ${var.project} && kubectl apply -f ../kubernetes/cert-manager/cert-manager.crds.yaml"
  }
}

// helm release
resource "helm_release" "cert-manager" {
  name             = "cert-manager"
  repository       = "https://charts.jetstack.io"
  chart            = "cert-manager"
  namespace        = kubernetes_namespace.cert-manager.metadata[0].name
  version          = "v0.15.1"
  depends_on = [null_resource.deploy-certmanager-crds]
}