provider "kubevirt" {
}

resource "kubevirt_virtual_machine" "myvm" {
  metadata {
    name = "myvm"
  }

  wait = true
  spec {
    running = true
    memory {
      request = "8Mi"
    }
    disks {
        name = "mydisk",
        disk {
            bus = "virtio"
        }
        volume {
          image = "kubevirt/cirros-registry-disk-demo"
        }
    }
  }
}
