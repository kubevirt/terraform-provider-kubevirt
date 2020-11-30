provider "kubevirt" {
}
provider "kubernetes" {
}

locals {
  namespace       = "terraform-provider-kubevirt-demo"
  vm_name_preffix = "test-vm"
  network_name    = "test-network"
  vm_count        = 2
}

// Create the ignition for the RHCOS startup

data "ignition_config" "vm_ignition_config" {
  count = local.vm_count

  users = [
    data.ignition_user.core.rendered,
  ]

  files = [
    element(data.ignition_file.hostname.*.rendered, count.index),
  ]
}

# Example configuration for the basic `core` user
data "ignition_user" "core" {
  name = "core"

  #Example password: foobar
  password_hash = "$5$XMoeOXG6$8WZoUCLhh8L/KYhsJN2pIRb3asZ2Xos3rJla.FA1TI7"
  # Preferably use the ssh key auth instead
  #ssh_authorized_keys = "${list()}"
}

# Replace the default hostname with our generated one
data "ignition_file" "hostname" {
  count = local.vm_count

  filesystem = "root" # default `ROOT` filesystem
  path       = "/etc/hostname"
  mode       = 420 # decimal 0644

  content {
    content = "${local.vm_name_preffix}-${count.index}"
  }
}

// Create the secret which holds the ignition data

resource "kubernetes_secret" "vm_ignition" {
  count = local.vm_count

  metadata {
    name      = "${local.vm_name_preffix}-${count.index}-ignition"
    namespace = local.namespace
  }
  data = {
    "userdata" = element(
      data.ignition_config.vm_ignition_config.*.rendered,
      count.index,
    )
  }
}

// Create the source data volume that all VMs should be cloned from

resource "kubevirt_data_volume" "data_volume" {
  metadata {
    name      = "source-dv"
    namespace = local.namespace
  }
  spec {
    source {
      http {
        url = "https://releases-art-rhcos.svc.ci.openshift.org/art/storage/releases/rhcos-4.4/44.81.202003062006-0/x86_64/rhcos-44.81.202003062006-0-openstack.x86_64.qcow2.gz"
      }
    }
    pvc {
      access_modes = ["ReadWriteMany"]
      resources {
        requests = {
          storage = "35Gi"
        }
      }
      storage_class_name = "standard"
    }
  }
}

resource "kubevirt_virtual_machine" "virtual_machine" {
  count = local.vm_count

  metadata {
    name      = "${local.vm_name_preffix}-${count.index}"
    namespace = local.namespace
    labels = {
      "key1" = "value1"
    }
  }
  spec {
    run_strategy = "Always"
    data_volume_templates {
      metadata {
        name      = "${local.vm_name_preffix}-bootvolume-${count.index}"
        namespace = local.namespace
      }
      spec {
        source {
          pvc {
            name      = kubevirt_data_volume.data_volume.metadata.0.name
            namespace = kubevirt_data_volume.data_volume.metadata.0.namespace
          }
        }
        pvc {
          access_modes = ["ReadWriteMany"]
          resources {
            requests = {
              storage = "35Gi"
            }
          }
          storage_class_name = "standard"
        }
      }
    }
    template {
      metadata {
        labels = {
          "kubevirt.io/vm" = "test-vm-${count.index}"
        }
      }
      spec {
        volume {
          name = "${local.vm_name_preffix}-datavolumedisk1-${count.index}"
          volume_source {
            data_volume {
              name = "${local.vm_name_preffix}-bootvolume-${count.index}"
            }
          }
        }
        volume {
          name = "${local.vm_name_preffix}-cloudinitdisk-${count.index}"
          volume_source {
            cloud_init_config_drive {
              user_data_secret_ref {
                name = kubernetes_secret.vm_ignition[count.index].metadata.0.name
              }
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
              name = "${local.vm_name_preffix}-datavolumedisk1-${count.index}"
              disk_device {
                disk {
                  bus = "virtio"
                }
              }
            }
            disk {
              name = "${local.vm_name_preffix}-cloudinitdisk-${count.index}"
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
              network_name = local.network_name
            }
          }
        }
      }
    }
  }
}
