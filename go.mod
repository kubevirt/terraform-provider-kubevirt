module github.com/kubevirt/terraform-provider-kubevirt

go 1.16

require (
	github.com/Azure/go-autorest/autorest v0.11.3 // indirect
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/aws/aws-sdk-go v1.31.9 // indirect
	github.com/golang/mock v1.4.3
	github.com/hashicorp/go-plugin v1.2.0
	github.com/hashicorp/hcl/v2 v2.6.0 // indirect
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/terraform v0.0.0-00010101000000-000000000000
	github.com/hashicorp/terraform-plugin-sdk v1.16.0
	github.com/mitchellh/cli v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
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
	github.com/hashicorp/terraform => github.com/openshift/terraform v0.12.20-openshift-4
	github.com/hashicorp/terraform-plugin-sdk => github.com/openshift/hashicorp-terraform-plugin-sdk v1.14.0-openshift
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
