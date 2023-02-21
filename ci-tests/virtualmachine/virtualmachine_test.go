package virtualmachine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/kubevirt/terraform-provider-kubevirt/ci-tests/common"
	"github.com/kubevirt/terraform-provider-kubevirt/ci-tests/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/pborman/uuid"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func TestVirtualMachine(t *testing.T) {
	format.MaxLength = 0
	RegisterFailHandler(Fail)
	RunSpecs(t, "Virtual Machine Suite")
}

var (
	testDir          string
	testID           string
	namespace        string
	vars             tfVars
	dvFromHttpSource cdiv1.DataVolumeSource
)

var _ = BeforeSuite(func() {
	var err error
	if testDir, err = ioutil.TempDir("", "virtualmachine-test-"); err != nil {
		Fail(fmt.Sprintf("failed to create temp dir for terraform execution, with error: %s", err))
	}
	testID = uuid.New()
	namespace = fmt.Sprintf("vm-test-namespace-%s", testID)
	common.CreateNamespace(namespace)
	vars = tfVars{
		VMName:    "test-vm",
		Namespace: namespace,
		URL:       "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2",
		Labels:    map[string]string{"key1": "value1"},
	}
	dvFromHttpSource = cdiv1.DataVolumeSource{
		HTTP: &cdiv1.DataVolumeSourceHTTP{
			URL: vars.URL,
		},
	}
})

var _ = AfterSuite(func() {
	common.DeleteNamespace(namespace)
	os.RemoveAll(testDir)
})

var _ = Describe("Virtual Machine Test", func() {
	testName := "virtualmachine"
	tfExecPath := "terraform"
	It("terraform init", func() {
		if err := terraform.Init(testDir, testName, tfExecPath); err != nil {
			Fail(fmt.Sprintf("failed to init terraform (runDir: %s, testName: %s, terraform path: %s) , with error: %s", testDir, testName, tfExecPath, err))
		}
	})
	It("terraform apply (create new)", func() {
		data, err := json.MarshalIndent(vars, "", "  ")
		if err != nil {
			Fail(fmt.Sprintf("failed to get data for tfvars file, with error: %s", err))
		}
		tfVarFiles := []*terraform.TfVarFile{
			{
				Filename: "terraform.auto.tfvars.json",
				Data:     data,
			},
		}
		var extraOpts []tfexec.ApplyOption
		if err = terraform.Apply(testDir, tfExecPath, tfVarFiles, extraOpts...); err != nil {
			Fail(fmt.Sprintf("failed to create virtual machine %s in namespace %s, with error: %s", vars.VMName, namespace, err))
		}
		common.ValidateVirtualMachine(vars.VMName, namespace, getExpectedVirtualMachine(vars.VMName, namespace, dvFromHttpSource, vars.Labels))
	})
	It("terraform apply (update)", func() {
		vars.Labels["key2"] = "value2"
		data, err := json.MarshalIndent(vars, "", "  ")
		if err != nil {
			Fail(fmt.Sprintf("failed to get data for tfvars file, with error: %s", err))
		}
		tfVarFiles := []*terraform.TfVarFile{
			{
				Filename: "terraform.auto.tfvars.json",
				Data:     data,
			},
		}
		var extraOpts []tfexec.ApplyOption
		if err = terraform.Apply(testDir, tfExecPath, tfVarFiles, extraOpts...); err != nil {
			Fail(fmt.Sprintf("failed to update Virtual Machine %s in namespace %s, with error: %s", vars.VMName, namespace, err))
		}
		common.ValidateVirtualMachine(vars.VMName, namespace, getExpectedVirtualMachine(vars.VMName, namespace, dvFromHttpSource, vars.Labels))
	})
	It("terraform destroy", func() {
		var extraOpts []tfexec.DestroyOption
		if err := terraform.Destroy(testDir, tfExecPath, extraOpts...); err != nil {
			Fail(fmt.Sprintf("failed to delete Virtual Machine %s in namespace %s, with error: %s", vars.VMName, namespace, err))
		}
		common.ValidateVirtualMachine(vars.VMName, namespace, nil)
	})
})

type tfVars struct {
	VMName    string            `json:"vm-name"`
	Namespace string            `json:"namespace"`
	URL       string            `json:"url"`
	Labels    map[string]string `json:"labels"`
}
