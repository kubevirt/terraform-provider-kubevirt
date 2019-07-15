provider "kubevirt" {
}

resource "kubevirt_virtual_machine" "myvm" {
  name = "myvm"
  namespace = "default"
  labels {
    label = "mylabel"
  }
  wait = true
  running = true
  image = {
    url = "http://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
    // url = "http://download.cirros-cloud.net/0.3.6/cirros-0.3.6-x86_64-disk.img"
  }
  memory {
    requests = "64M"
    limits = "256M"
  }
  cpu {
    requests = "100m"
    limits = "200m"
    cores = 1
    threads = 2
  }
  interfaces {
    name = "nic1"
    type = "bridge"
    network = "pod"
  }
  cloud_init = <<-EOF
  #cloud-config
  password: fedora
  chpasswd: { expire: False }
  EOF
}
