package test_utils

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func NullifySchemaSetFunction(ss *schema.Set) {
	ss.F = nil
}

func GetDataVolume(vm interface{}) interface{} {
	return vm.(map[string]interface{})["data_volume_templates"].([]interface{})[0]
}

func GetPVCRequirements(dataVolume interface{}) interface{} {
	return dataVolume.(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["pvc"].([]interface{})[0].(map[string]interface{})["resources"].([]interface{})[0]
}

func GetDomainResources(vm interface{}) interface{} {
	return vm.(map[string]interface{})["template"].([]interface{})[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["domain"].([]interface{})[0].(map[string]interface{})["resources"].([]interface{})[0]
}

func GetVirtualMachineTolerations(vm interface{}) interface{} {
	return vm.(map[string]interface{})["template"].([]interface{})[0].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})["tolerations"].([]interface{})[0]
}