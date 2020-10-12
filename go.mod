module github.com/kubevirt/terraform-provider-kubevirt

go 1.14

require (
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/terraform v0.13.4
	github.com/mitchellh/go-homedir v1.1.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v10.0.0+incompatible
)

replace (
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
