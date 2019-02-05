provider "kubevirt" {
}

resource "kubevirt_virtual_machine" "myvm" {
  metadata {
    name = "myvm"
  }

  spec {
    running = true
    memory = "8Mi"
    disks {
        name = "mydisk",
        bus = "virtio",
        volume {
          image = "kubevirt/cirros-registry-disk-demo"
        }
    }
  }
}
