terraform {
  required_version = ">= 0.13"

  required_providers {
    helm = {
      version = ">= 2.5.1"
      source  = "hashicorp/helm"
    }
    k8s = {
      version = "0.9.1"
      source  = "banzaicloud/k8s"
    }
    kubernetes = {
      version = ">= 2.0.2"
      source  = "hashicorp/kubernetes"
    }
  }
}