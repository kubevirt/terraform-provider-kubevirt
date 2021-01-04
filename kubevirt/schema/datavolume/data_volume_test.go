package datavolume

import (
	"testing"

	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/test_utils/expand_utils"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/test_utils/flatten_utils"
	"gotest.tools/assert"

	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/test_utils"
)


func TestExpandDataVolumeTemplates(t *testing.T) {
	baseOutput := expand_utils.GetBaseOutputForDataVolume()

	cases := []struct {
		name string
		shouldError bool
		expectedOutput []cdiv1.DataVolume
		expectedErrorMessage string
		modifier func(interface{})
	}{
		{
			name: "working case",
			shouldError: false,
			expectedOutput: []cdiv1.DataVolume{
				baseOutput,
			},
		},
		{
			name: "bad pvc requests",
			shouldError: true,
			modifier: func(input interface{}){
				pvcRequirements := test_utils.GetPVCRequirements(input)
				pvcRequirements.(map[string]interface{})["requests"].(map[string]interface{})["storage"] = "a5"
			},
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
		{
			name: "bad pvc limits",
			shouldError: true,
			modifier: func(input interface{}){
				pvcRequirements := test_utils.GetPVCRequirements(input)
				pvcRequirements.(map[string]interface{})["limits"].(map[string]interface{})["storage"] = "a5"
			},
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := expand_utils.GetBaseInputForDataVolume()

			if (tc.modifier != nil) {
				tc.modifier(input)
			}
			output, err := ExpandDataVolumeTemplates([]interface{}{input})

			if (tc.shouldError) {
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, output[0], baseOutput)	
			}
		})
	}
}

func TestFlattenDataVolumeTemplates(t *testing.T) {
	input1 := flatten_utils.GetBaseInputForDataVolume()
	output1 := flatten_utils.GetBaseOutputForDataVolume()

	cases := []struct {
		Input          []cdiv1.DataVolume
		ExpectedOutput []interface{}
	} {
		{
			Input: []cdiv1.DataVolume{
				input1,
			},
			ExpectedOutput: []interface{}{
				output1,
			},
		},
	}

	for _, tc := range cases {
		output := FlattenDataVolumeTemplates(tc.Input)

		//Some fields include terraform randomly generated params that can't be compared
		//so we need to manually remove them
		nullifyUncomparableFields(&output)
		nullifyUncomparableFields(&tc.ExpectedOutput)

		assert.DeepEqual(t, output, tc.ExpectedOutput)
	}
}

func nullifyUncomparableFields(output *[]interface{}) {
	accessModes := (*output)[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["pvc"].([]interface{})[0].(map[string]interface{})["access_modes"]

	test_utils.NullifySchemaSetFunction(accessModes.(*schema.Set))
}
