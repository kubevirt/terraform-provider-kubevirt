provider "kubevirt" {
}
provider "kubernetes" {
}

resource "kubevirt_virtual_machine" "virtual_machine" {
  metadata {
    name      = "test-vm"
    namespace = "tenantcluster"
    labels = {
      "key1" = "value1"
    }
  }
  spec {
    run_strategy = "Always"
    data_volume_templates {
      metadata {
        name      = "test-vm-bootvolume"
        namespace = "tenantcluster"
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
    template {
      metadata {
        labels = {
          "kubevirt.io/vm" = "test-vm"
        }
      }
      spec {
        volume {
          name = "test-vm-datavolumedisk1"
          volume_source {
            data_volume {
              name = "test-vm-bootvolume"
            }
          }
        }
        domain {
          resources {
            requests = {
              memory = "10G"
              cpu    = 4
            }
          }
          devices {
            disk {
              name = "test-vm-datavolumedisk1"
              disk_device {
                disk {
                  bus = "virtio"
                }
              }
            }
            interface {
              name                     = "main"
              interface_binding_method = "InterfaceBridge"
            }
          }
        }
        network {
          name = "main"
          network_source {
            multus {
              network_name = "tenantcluster"
            }
          }
        }
        affinity {
          pod_anti_affinity {
            preferred_during_scheduling_ignored_during_execution {
              weight = 100
              pod_affinity_term {
                label_selector {
                  match_labels = {
                    anti-affinity-key = "anti-affinity-val"
                  }
                }
                topology_key = "kubernetes.io/hostname"
              }
            }
          }
        }
      }
    }
  }
}
