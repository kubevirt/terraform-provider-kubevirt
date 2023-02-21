package datavolume

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

func TestDataVolume(t *testing.T) {
	format.MaxLength = 0
	RegisterFailHandler(Fail)
	RunSpecs(t, "Data Volume Suite")
}

var (
	testDir   string
	testID    string
	namespace string
	vars      tfVars
)

var _ = BeforeSuite(func() {
	var err error
	if testDir, err = ioutil.TempDir("", "datavolume-test-"); err != nil {
		Fail(fmt.Sprintf("failed to create temp dir for terraform execution, with error: %s", err))
	}
	testID = uuid.New()
	namespace = fmt.Sprintf("datavolume-test-namespace-%s", testID)
	common.CreateNamespace(namespace)
	vars = tfVars{
		DvFromHttpName: "test-dv-from-http",
		DvFromPVCName:  "test-dv-from-pvc",
		Namespace:      namespace,
		URL:            "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2",
		Labels:         map[string]string{"key1": "value1"},
	}
})

var _ = AfterSuite(func() {
	common.DeleteNamespace(namespace)
	os.RemoveAll(testDir)
})

var _ = Describe("Data Volume Test", func() {
	testName := "datavolume"
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
			Fail(fmt.Sprintf("failed to create data volumes [%s, %s] in namespace %s, with error: %s", vars.DvFromHttpName, vars.DvFromPVCName, namespace, err))
		}
		validateDVs(vars)
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
			Fail(fmt.Sprintf("failed to update data volumes [%s, %s] in namespace %s, with error: %s", vars.DvFromHttpName, vars.DvFromPVCName, namespace, err))
		}
		validateDVs(vars)
	})
	It("terraform destroy", func() {
		var extraOpts []tfexec.DestroyOption
		if err := terraform.Destroy(testDir, tfExecPath, extraOpts...); err != nil {
			Fail(fmt.Sprintf("failed to delete data volumes [%s, %s] in namespace %s, with error: %s", vars.DvFromHttpName, vars.DvFromPVCName, namespace, err))
		}
		common.ValidateDatavolume(vars.DvFromHttpName, namespace, nil)
		common.ValidateDatavolume(vars.DvFromPVCName, namespace, nil)
	})
})

func validateDVs(vars tfVars) {
	// validate data volume created from http source
	dvFromHttpSource := cdiv1.DataVolumeSource{
		HTTP: &cdiv1.DataVolumeSourceHTTP{
			URL: vars.URL,
		},
	}
	common.ValidateDatavolume(vars.DvFromHttpName, namespace, getExpectedDataVolume(vars.DvFromHttpName, namespace, dvFromHttpSource, vars.Labels))
	// validate data volume created from http source
	dvFromPVCSource := cdiv1.DataVolumeSource{
		PVC: &cdiv1.DataVolumeSourcePVC{
			Name:      vars.DvFromHttpName,
			Namespace: namespace,
		},
	}
	common.ValidateDatavolume(vars.DvFromPVCName, namespace, getExpectedDataVolume(vars.DvFromPVCName, namespace, dvFromPVCSource, vars.Labels))
}

type tfVars struct {
	DvFromHttpName string            `json:"dv-from-http-name"`
	DvFromPVCName  string            `json:"dv-from-pvc-name"`
	Namespace      string            `json:"namespace"`
	URL            string            `json:"url"`
	Labels         map[string]string `json:"labels"`
}
