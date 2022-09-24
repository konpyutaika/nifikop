# Create Zookeeper namespace if asked
resource "kubernetes_namespace" "zookeeper" {
  count = var.zookeeper_create_namespace ? 1 : 0
  metadata {
    annotations = {
      name = var.zookeeper_namespace
    }
    name = var.zookeeper_namespace
  }
}

# Create service account
resource "kubernetes_service_account" "zookeeper" {
  metadata {
    name      = "zookeeper"
    namespace = var.zookeeper_namespace
  }
  automount_service_account_token = true

  depends_on = [
    kubernetes_namespace.zookeeper
  ]
}

resource "kubernetes_pod_security_policy" "zookeeper" {
  metadata {
    name = "zookeeper"
  }
  spec {
    privileged                 = false
    allow_privilege_escalation = false
    allowed_unsafe_sysctls     = ["vm.max_map_count"]

    default_allow_privilege_escalation = false
    forbidden_sysctls = [
      "kernel.*",
      "net.*",
      "dev.*",
      "fs.*",
    ]

    fs_group {
      rule = "MustRunAs"
      range {
        min = 1
        max = 65535
      }
    }

    supplemental_groups {
      rule = "MustRunAs"
      range {
        min = 1
        max = 65535
      }
    }

    required_drop_capabilities = [
      "AUDIT_CONTROL",
      "AUDIT_READ",
      "AUDIT_WRITE",
      "BLOCK_SUSPEND",
      "CHOWN",
      "DAC_OVERRIDE",
      "DAC_READ_SEARCH",
      "FOWNER",
      "FSETID",
      "IPC_LOCK",
      "IPC_OWNER",
      "KILL",
      "LEASE",
      "LINUX_IMMUTABLE",
      "MAC_ADMIN",
      "MAC_OVERRIDE",
      "MKNOD",
      "NET_ADMIN",
      "NET_BIND_SERVICE",
      "NET_BROADCAST",
      "NET_RAW",
      "SETGID",
      "SETFCAP",
      "SETPCAP",
      "SETUID",
      "SYS_ADMIN",
      "SYS_BOOT",
      "SYS_CHROOT",
      "SYS_MODULE",
      "SYS_NICE",
      "SYS_PACCT",
      "SYS_PTRACE",
      "SYS_RAWIO",
      "SYS_TIME",
      "SYS_TTY_CONFIG",
      "SYSLOG",
      "WAKE_ALARM",
    ]

    se_linux {
      rule = "RunAsAny"
    }

    run_as_user {
      rule = "MustRunAs"
      range {
        min = 1
        max = 65535
      }
    }

    volumes = [
      "configMap",
      "emptyDir",
      "projected",
      "secret",
      "downwardAPI",
      "persistentVolumeClaim",
    ]
  }
}

resource "kubernetes_cluster_role" "zookeeper" {
  metadata {
    name = "zookeeper-psp"
  }

  rule {
    api_groups     = ["extensions", ]
    resources      = ["podsecuritypolicies", ]
    verbs          = ["use", ]
    resource_names = [kubernetes_pod_security_policy.zookeeper.metadata[0].name]
  }
}

# Binding external-dns cluster role, with the external-dns Service account.
resource "kubernetes_cluster_role_binding" "zookeeper" {
  metadata {
    name = "zoookeeper-psp-${var.zookeeper_namespace}"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.zookeeper.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.zookeeper.metadata[0].name
    namespace = var.zookeeper_namespace
  }

  depends_on = [
    kubernetes_namespace.zookeeper
  ]
}

# Network policy
resource "kubernetes_network_policy" "zookeeper" {
  metadata {
    name      = "zookeeper-ingress-egress"
    namespace = var.zookeeper_namespace
  }
  spec {
    policy_types = ["Ingress", "Egress"]
    pod_selector {
    }

    ingress {}

    egress {}
  }

  depends_on = [
    kubernetes_namespace.zookeeper
  ]
}

# helm release
resource "helm_release" "zookeeper" {
  name       = "zookeeper"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "zookeeper"
  namespace  = var.zookeeper_namespace
  version    = "9.2.2"

  dynamic "set" {
    for_each = var.zookeeper_config
    content {
      name  = set.key
      value = set.value
    }
  }

  set {
    name  = "serviceAccount.create"
    value = "false"
  }

  set {
    name  = "serviceAccount.name"
    value = kubernetes_service_account.zookeeper.metadata[0].name
  }

  set {
    name  = "global.storageClass"
    value = "ssd-wait"
  }

  depends_on = [
    kubernetes_namespace.zookeeper
  ]
}