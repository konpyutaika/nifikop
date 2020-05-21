resource "kubernetes_namespace" "cert-manager" {
  metadata {
    annotations = {
      name = var.cert_manager_namespace
    }
    # Enable istio sidecar injection, into pods instantiate into this namespace.
    labels = {
      istio-injection = "enabled"
      istio-operator-managed-injection = "enabled"
    }
    name = var.cert_manager_namespace
  }
  depends_on = [google_container_node_pool.nodes]
}

resource "null_resource" "deploy-certmanager-crds" {
  depends_on = [kubernetes_namespace.cert-manager]
  provisioner "local-exec" {
    command = "gcloud container clusters get-credentials ${google_container_cluster.source-squidflow-cluster.name}  --zone ${var.zone} --project ${var.project} && kubectl apply -f ../kubernetes/cert-manager"
  }
}

// helm release
resource "helm_release" "cert-manager" {
  name             = "cert-manager"
  repository       = data.helm_repository.jetstack.metadata[0].name
  chart            = "jetstack/cert-manager"
  namespace        = kubernetes_namespace.cert-manager.metadata[0].name
  version          = "v0.11.0"
  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding, /*helm_release.istio-operator,*/ null_resource.deploy-certmanager-crds]
}