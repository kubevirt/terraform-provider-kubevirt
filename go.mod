module github.com/kubevirt/terraform-provider-kubevirt

go 1.14

require (
	github.com/hashicorp/terraform v0.13.4
	github.com/mitchellh/go-homedir v1.1.0
	k8s.io/api v0.19.0
	k8s.io/apiextensions-apiserver v0.19.0 // indirect
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v12.0.0+incompatible
	kubevirt.io/client-go v0.29.0
	kubevirt.io/containerized-data-importer v1.10.6
)

replace (
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
