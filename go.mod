module github.com/kubevirt/terraform-provider-kubevirt

go 1.14

require (
	github.com/Azure/go-autorest/autorest v0.11.3 // indirect
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/aws/aws-sdk-go v1.31.9 // indirect
	github.com/golang/mock v1.4.3
	github.com/hashicorp/hcl/v2 v2.6.0 // indirect
	github.com/hashicorp/terraform-config-inspect v0.0.0-20191212124732-c6ae6269b9d7 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.16.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/zclconf/go-cty v1.5.1 // indirect
	github.com/zclconf/go-cty-yaml v1.0.2 // indirect
	gotest.tools v2.2.0+incompatible
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
