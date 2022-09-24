# Create cert manager namespace if asked
resource "kubernetes_namespace" "cert_manager" {
  count = var.cert_manager_create_namespace ? 1 : 0
  metadata {
    annotations = {
      name = var.cert_manager_namespace
    }
    name = var.cert_manager_namespace
  }
}

# Deploy CRDs
resource "k8s_manifest" "cert_manager_crds" {
  for_each = fileset("${path.module}/kubernetes/cert-manager/crds/${var.cert_manager_version}/", "*")
  content  = file("${path.module}/kubernetes/cert-manager/crds/${var.cert_manager_version}/${each.value}")
}

# helm release
resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  namespace  = kubernetes_namespace.cert_manager[0].metadata[0].name
  version    = var.cert_manager_version

  dynamic "set" {
    for_each = var.cert_manager_config
    content {
      name  = set.key
      value = set.value
    }
  }

  depends_on = [k8s_manifest.cert_manager_crds]
}
