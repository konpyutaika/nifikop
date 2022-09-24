locals {
  nifi_registry_name     = "nifi-registry"
  nifi_registry_svc_port = 80
}

resource "kubernetes_service_account" "nifi_registry" {
  count = var.enable_nifi_registry ? 1 : 0
  metadata {
    name        = "nifi-registry"
    namespace   = var.nifi_registry_namespace
    annotations = var.nifi_registry_sa_annotations
  }
  automount_service_account_token = true

  depends_on = [
    kubernetes_namespace.nifikop
  ]
}

resource "kubernetes_secret" "nifi_registry_secret" {
  count = var.enable_nifi_registry ? 1 : 0
  metadata {
    name      = "nifi-registry-secret"
    namespace = var.nifi_registry_namespace
  }

  data = {
    "ssh-key"               = "%{if var.nifi_registry_git_config.ssh_key_path != ""}${base64encode(file(var.nifi_registry_git_config.ssh_key_path))}%{else}%{endif}"
    "db-pass"               = var.nifi_registry_database_config.password
    "db-ssl-cert"           = var.nifi_registry_database_ssl_config.cert
    "db-ssl-private-key"    = var.nifi_registry_database_ssl_config.private_key
    "db-ssl-server-ca-cert" = var.nifi_registry_database_ssl_config.server_ca_cert
  }

  depends_on = [
    kubernetes_namespace.nifikop
  ]
}

locals {
  deployment_manifest = templatefile("${path.module}/kubernetes/nifi-registry/deployment.yaml.tpl", {
    // Deployment configuration
    name                    = local.nifi_registry_name
    backend                 = var.nifi_registry_backend
    namespace               = var.nifi_registry_namespace
    service-account-name    = var.enable_nifi_registry ? kubernetes_service_account.nifi_registry[0].metadata[0].name : ""
    container-image         = var.nifi_registry_image
    container-port          = var.nifi_registry_container_port
    node-selector-node-pool = var.nifi_registry_node_selector_node_pool

    secret-name = var.enable_nifi_registry ? kubernetes_secret.nifi_registry_secret[0].metadata[0].name : ""

    // External Git storage
    git-config-user-email = var.nifi_registry_git_config.user_email
    git-remote-url        = var.nifi_registry_git_config.remote_url
    git-remote-branch     = var.nifi_registry_git_config.remote_branch
    git-remote-to-push    = var.nifi_registry_git_config.remote_to_push
    ssh-known-hosts       = "%{if var.nifi_registry_git_config.ssh_known_hosts_path != ""}${base64encode(file(var.nifi_registry_git_config.ssh_known_hosts_path))}%{else}%{endif}"

    // External DB storage
    db-url   = var.nifi_registry_database_config.url
    db-class = var.nifi_registry_database_config.driver_class
    db-user  = var.nifi_registry_database_config.user

    sidecars = flatten([for sidecar in var.nifi_registry_sidecars : indent(10,yamlencode(sidecar))])

  })

  svc_manifest = templatefile("${path.module}/kubernetes/nifi-registry/svc.yaml.tpl", {
    annotations  = var.nifi_registry_svc_annotations
    namespace    = var.nifi_registry_namespace
    target-port  = var.nifi_registry_container_port
    app-label    = local.nifi_registry_name
    service-type = var.nifi_registry_service_type
    port         = local.nifi_registry_svc_port
  })
}

resource "k8s_manifest" "nifi_registry_deployment" {
  count = var.enable_nifi_registry ? 1 : 0
  content = local.deployment_manifest
  depends_on = [
    kubernetes_service_account.nifi_registry,
    kubernetes_secret.nifi_registry_secret
  ]
}

resource "k8s_manifest" "nifi_registry_svc" {
  count      = var.enable_nifi_registry ? 1 : 0
  content    = local.svc_manifest
  depends_on = [
    k8s_manifest.nifi_registry_deployment
  ]
}
