provider "kubevirt" {

}
provider "kubernetes" {

}

variable "minikube_ip" {}

resource "kubevirt_virtual_machine" "myvm" {
  metadata {
    name = "myvm"
    labels {
      vm = "myvm"
    }
  }

  wait = true
  spec {
    running = true

    disks {
      name = "mydisk",
      disk {
        bus = "virtio"
      }
      volume {
        image = "kubevirt/cirros-registry-disk-demo"
      }
    }

    memory {
      request = "64Mi"
    }
  }
}

resource "kubernetes_service" "myvmservice" {
  metadata {
    name = "myvmservice"
  }
  spec {
    selector {
      vm = "${kubevirt_virtual_machine.myvm.metadata.0.labels.vm}"
    }
    session_affinity = "ClientIP"
    port {
      name = "ssh"
      node_port = 30000
      port = 27017
      target_port = 22
    }

    type = "NodePort"
  }

  provisioner "file" {
    content = "Test\n"
    destination = "/tmp/test"
    connection {
      type     = "ssh"
      user     = "cirros"
      password = "gocubsgo"
      port     = 30000
      host     = "${var.minikube_ip}"
    }
  }
}
