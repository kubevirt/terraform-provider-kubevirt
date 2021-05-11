provider "kubevirt" {
}

resource "kubevirt_virtual_machine" "virtual_machine" {
  metadata {
    name      = var.vm-name
    namespace = var.namespace
    labels    = var.labels
  }
  spec {
    run_strategy = "Always"
    data_volume_templates {
      metadata {
        name      = "${var.vm-name}-bootvolume"
        namespace = var.namespace
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
    template {
      metadata {
        namespace = var.namespace
        labels    = { "templateKey1" = "templateVal1" }
      }
      spec {
        volume {
          name = "datavolumedisk1"
          volume_source {
            data_volume {
              name = "${var.vm-name}-bootvolume"
            }
          }
        }
        domain {
          resources {
            requests = {
              memory = "120Mi"
              cpu    = 1
            }
          }
          devices {
            disk {
              name = "datavolumedisk1"
              disk_device {
                disk {
                  bus = "virtio"
                }
              }
            }
            interface {
              name                     = "default"
              interface_binding_method = "InterfaceBridge"
            }
          }
        }
        network {
          name = "default"
          network_source {
            pod {}
          }
        }
      }
    }
  }
}
