package virtualmachine

import (
	"testing"

	kubevirtapiv1 "kubevirt.io/client-go/api/v1"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/test_utils"

	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/test_utils/expand_utils"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/test_utils/flatten_utils"

	"gotest.tools/assert"
)

func TestExpandVirtualMachineSpec(t *testing.T) {
	baseOutput := expand_utils.GetBaseOutputForVirtualMachine()

	cases := []struct {
		name string
		shouldError bool
		expectedOutput []kubevirtapiv1.VirtualMachineSpec
		expectedErrorMessage string
		modifier func(interface{})
	}{
		{
			name: "working case",
			shouldError: false,
			expectedOutput: []kubevirtapiv1.VirtualMachineSpec{
				baseOutput,
			},
		},
		{
			name: "bad toleration_seconds",
			shouldError: true,
			modifier: func(input interface{}){
				tolerations := test_utils.GetVirtualMachineTolerations(input)
				tolerations.(map[string]interface{})["toleration_seconds"] = "a5"
			},
			expectedErrorMessage: "invalid toleration_seconds must be int or \"\", got \"a5\"",
		},
		{
			name: "bad pvc requests",
			shouldError: true,
			modifier: func(input interface{}){
				pvcRequirements := test_utils.GetPVCRequirements(test_utils.GetDataVolume(input))
				pvcRequirements.(map[string]interface{})["requests"].(map[string]interface{})["storage"] = "a5"
			},
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
		{
			name: "bad pvc limits",
			shouldError: true,
			modifier: func(input interface{}){
				pvcRequirements := test_utils.GetPVCRequirements(test_utils.GetDataVolume(input))
				pvcRequirements.(map[string]interface{})["limits"].(map[string]interface{})["storage"] = "a5"
			},
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
		{
			name: "bad domain resource requests",
			shouldError: true,
			modifier: func(input interface{}){
				domainResources := test_utils.GetDomainResources(input)
				domainResources.(map[string]interface{})["requests"].(map[string]interface{})["storage"] = "a5"
			},
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
		{
			name: "bad domain resource limits",
			shouldError: true,
			modifier: func(input interface{}){
				domainResources := test_utils.GetDomainResources(input)
				domainResources.(map[string]interface{})["limits"].(map[string]interface{})["storage"] = "a5"
			},
			expectedErrorMessage: "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := expand_utils.GetBaseInputForVirtualMachine()

			if (tc.modifier != nil) {
				tc.modifier(input)
			}
			output, err := expandVirtualMachineSpec([]interface{}{input})

			if (tc.shouldError) {
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, output, baseOutput)	
			}
		})
	}
}

func TestFlattenVirtualMachineSpec(t *testing.T) {
	input1 := flatten_utils.GetBaseInputForVirtualMachine()
	output1 := flatten_utils.GetBaseOutputForVirtualMachine()

	cases := []struct {
		// name string
		input          kubevirtapiv1.VirtualMachineSpec
		// shouldError bool
		expectedOutput []interface{}
	}{
		{
			input: input1,
			expectedOutput: []interface{}{
				output1,
			},
		},
	}

	for _, tc := range cases {
		output := flattenVirtualMachineSpec(tc.input)

		//Some fields include terraform randomly generated params that can't be compared
		//so we need to manually remove them
		nullifyUncomparableFields(&output)
		nullifyUncomparableFields(&tc.expectedOutput)

		assert.DeepEqual(t, output, tc.expectedOutput)
	}
}

func nullifyUncomparableFields(output *[]interface{}) {
	accessModes := (*output)[0].(map[string]interface{})["data_volume_templates"].([]interface{})[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["pvc"].([]interface{})[0].(map[string]interface{})["access_modes"]
	test_utils.NullifySchemaSetFunction(accessModes.(*schema.Set))

	vmAffinity := (*output)[0].(map[string]interface{})["template"].([]interface{})[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["affinity"]

	podAntiAffinity := vmAffinity.([]interface{})[0].(map[string]interface{})["pod_anti_affinity"].([]interface{})[0].(map[string]interface{})
	
	podAntiAffinityPreferredNamespace := podAntiAffinity["preferred_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["pod_affinity_term"].([]interface{})[0].(map[string]interface{})["namespaces"]
	test_utils.NullifySchemaSetFunction(podAntiAffinityPreferredNamespace.(*schema.Set))

	podAntiAffinityRequiredNamespace := podAntiAffinity["required_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["namespaces"]
	test_utils.NullifySchemaSetFunction(podAntiAffinityRequiredNamespace.(*schema.Set))

	podAffinity := vmAffinity.([]interface{})[0].(map[string]interface{})["pod_affinity"].([]interface{})[0].(map[string]interface{})

	podAffinityPreferredNamespace := podAffinity["preferred_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["pod_affinity_term"].([]interface{})[0].(map[string]interface{})["namespaces"]
	test_utils.NullifySchemaSetFunction(podAffinityPreferredNamespace.(*schema.Set))

	podAffinityRequiredNamespace := podAffinity["required_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["namespaces"]
	test_utils.NullifySchemaSetFunction(podAffinityRequiredNamespace.(*schema.Set))

	nodeAffinity := vmAffinity.([]interface{})[0].(map[string]interface{})["node_affinity"].([]interface{})[0].(map[string]interface{})

	nodeSelector := nodeAffinity["required_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["node_selector_term"].([]interface{})[0].(map[string]interface{})

	nodeRequiredMatchExpressions := nodeSelector["match_expressions"].([]interface{})[0].(map[string]interface{})["values"]
	test_utils.NullifySchemaSetFunction(nodeRequiredMatchExpressions.(*schema.Set))

	nodeRequiredMatchFields := nodeSelector["match_fields"].([]interface{})[0].(map[string]interface{})["values"]
	test_utils.NullifySchemaSetFunction(nodeRequiredMatchFields.(*schema.Set))

	nodePreference := nodeAffinity["preferred_during_scheduling_ignored_during_execution"].([]interface{})[0].(map[string]interface{})["preference"].([]interface{})[0].(map[string]interface{})

	nodePreferredMatchExpressions := nodePreference["match_expressions"].([]interface{})[0].(map[string]interface{})["values"]
	test_utils.NullifySchemaSetFunction(nodePreferredMatchExpressions.(*schema.Set))

	nodePreferredMatchFields := nodePreference["match_fields"].([]interface{})[0].(map[string]interface{})["values"]
	test_utils.NullifySchemaSetFunction(nodePreferredMatchFields.(*schema.Set))
}
