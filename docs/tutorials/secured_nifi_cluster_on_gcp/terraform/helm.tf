#################################
#       Helm : Tiller SA        #
#################################
// Bind state of tiller's secret
resource "kubernetes_secret" "tiller" {
  metadata {
    name = "tiller"
    namespace = "kube-system"
  }
}

// Create tiller service account on Kubernetes
resource "kubernetes_service_account" "tiller" {
  metadata {
    name = "tiller"
    namespace = kubernetes_secret.tiller.metadata.0.namespace
  }
  depends_on = [google_container_node_pool.nodes]
}

// Bind tiller service account with cluster role admin on K8S
resource "kubernetes_cluster_role_binding" "tiller-admin-binding" {
  metadata {
    name      = "tiller-admin-binding"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.tiller.metadata.0.name
    namespace = kubernetes_service_account.tiller.metadata.0.namespace
  }
  depends_on = [kubernetes_service_account.tiller]
}

// helm repository
data "helm_repository" "jetstack" {
  name = "jetstack"
  url  = "https://charts.jetstack.io"

  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding]
}

// helm repository
data "helm_repository" "orange-incubator" {
  name = "orange-incubator"
  url  = "https://orange-kubernetes-charts-incubator.storage.googleapis.com"
  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding]
}

// helm repository
data "helm_repository" "bitnami" {
  name = "bitnami"
  url  = "https://charts.bitnami.com/bitnami"

  depends_on = [kubernetes_cluster_role_binding.tiller-admin-binding]
}