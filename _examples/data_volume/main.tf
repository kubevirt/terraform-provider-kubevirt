provider "kubernetes" {
}

provider "kubevirt" {
}

// This pvc is created as a source for "dv from pvc" example
resource "kubernetes_persistent_volume_claim" "example" {
  metadata {
    name      = "exampleclaimname"
    namespace = "test-terraform-provider"
  }
  spec {
    access_modes = ["ReadWriteMany"]
    resources {
      requests = {
        storage = "5Gi"
      }
    }
  }
}

resource "kubevirt_data_volume" "data_volume_pvc" {
  metadata {
    name      = "dv-example-clone-pvc"
    namespace = "test-terraform-provider"
  }
  spec {
    source {
      pvc {
        name      = kubernetes_persistent_volume_claim.example.metadata.0.name
        namespace = kubernetes_persistent_volume_claim.example.metadata.0.namespace
      }
    }
    pvc {
      access_modes = ["ReadWriteOnce"]
      resources {
        requests = {
          storage = "10Gi"
        }
      }
      storage_class_name = "standard"
    }
  }
}

resource "kubevirt_data_volume" "data_volume_http" {
  metadata {
    name      = "data-volume-from-http"
    namespace = "test-terraform-provider"
    labels = {
      "key1" = "value1"
    }
  }
  spec {
    source {
      http {
        url = "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2"
      }
    }
    pvc {
      access_modes = ["ReadWriteOnce"]
      resources {
        requests = {
          storage = "10Gi"
        }
      }
      storage_class_name = "standard"
    }
  }
}
