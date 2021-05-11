provider "kubevirt" {
}

resource "kubevirt_data_volume" "data_volume_http" {
  metadata {
    name      = var.dv-from-http-name
    namespace = var.namespace
    labels    = var.labels
  }
  spec {
    source {
      http {
        url = var.url
      }
    }
    pvc {
      access_modes = ["ReadWriteOnce"]
      resources {
        requests = {
          storage = "10Gi"
        }
      }
    }
  }
}

resource "kubevirt_data_volume" "data_volume_pvc" {
  metadata {
    name      = var.dv-from-pvc-name
    namespace = var.namespace
    labels    = var.labels
  }
  spec {
    source {
      pvc {
        name      = kubevirt_data_volume.data_volume_http.metadata.0.name
        namespace = kubevirt_data_volume.data_volume_http.metadata.0.namespace
      }
    }
    pvc {
      access_modes = ["ReadWriteOnce"]
      resources {
        requests = {
          storage = "10Gi"
        }
      }
    }
  }
}
