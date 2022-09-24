output "nifikop_namespace" {
  description = "NiFiKop namespace"
  value       = var.nifikop_namespace
}

output "nifikop_name" {
  description = "NiFiKop names"
  value       = helm_release.nifikop.name

}

output "nifikop_image_tag" {
  description = "The version of the NiFiKop instance"
  value       = var.nifikop_image_tag
}

output "nifi_registry_name" {
  description = "NiFi registry name"
  value       = local.nifi_registry_name
  depends_on = [
    k8s_manifest.nifi_registry_deployment
  ]
}

output "nifi_registry_namespace" {
  description = "NiFi registry namespace"
  value       = var.nifi_registry_namespace
  depends_on = [
    k8s_manifest.nifi_registry_deployment
  ]
}

output "nifi_registry_container_port" {
  description = "NiFi registry container port"
  value       = var.nifi_registry_container_port
  depends_on = [
    k8s_manifest.nifi_registry_deployment
  ]
}

output "nifi_registry_service_port" {
  description = "NiFi registry service port"
  value       = local.nifi_registry_svc_port
  depends_on = [
    k8s_manifest.nifi_registry_svc
  ]
}

output "zookeeper_namespace" {
  description = "Zookeeper namespace"
  value       = helm_release.zookeeper.namespace
}

output "zookeeper_name" {
  description = "Zookeeper name"
  value       = helm_release.zookeeper.name
}
