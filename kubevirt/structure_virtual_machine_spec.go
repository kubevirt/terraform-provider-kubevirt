package kubevirt

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func expandVirtualMachineSpec(l []interface{}) (*map[string]interface{}, error) {
	obj := make(map[string]interface{})
	if len(l) == 0 || l[0] == nil {
		return &obj, nil
	}
	spec := l[0].(map[string]interface{})
	obj["spec"] = map[string]interface{}{
		"running": spec["running"].(bool),
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"domain": map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]string{
							"memory": spec["memory"].(string),
						},
					},
					"devices": map[string]interface{}{
						"disks": expandVirtualMachineDisksSpec(spec["disks"].([]interface{})),
					},
				},
				"volumes": expandVirtualMachineVolumesSpec(spec["disks"].([]interface{})),
			},
		},
	}

	return &obj, nil
}

func expandVirtualMachineDisksSpec(d []interface{}) []map[string]interface{} {
	disks := make([]map[string]interface{}, len(d))
	for i, v := range d {
		disks[i] = map[string]interface{}{
			"name": v.(map[string]interface{})["name"].(string),
			"disk": map[string]string{
				"bus": v.(map[string]interface{})["bus"].(string),
			},
		}
	}
	return disks
}

func expandVirtualMachineVolumesSpec(d []interface{}) []map[string]interface{} {
	volumes := make([]map[string]interface{}, len(d))
	for i, v := range d {
		vv := v.(map[string]interface{})
		volume := vv["volume"].(map[string]interface{})
		volumeimage := volume["image"].(string)

		volumes[i] = map[string]interface{}{
			"name": vv["name"].(string),
			"containerDisk": map[string]string{
				"image": volumeimage,
			},
		}
	}
	return volumes
}

func flattenVMMetadata(meta map[string]interface{}) []map[string]interface{} {
	m := make(map[string]interface{})
	m["annotations"] = meta["annotations"]
	if meta["generateName"] != "" {
		m["generate_name"] = meta["generateName"]
	}
	m["labels"] = meta["labels"]
	m["name"] = meta["name"]
	m["resource_version"] = meta["resourceVersion"]
	m["self_link"] = meta["selfLink"]
	m["uid"] = fmt.Sprintf("%v", meta["uid"])
	m["generation"] = meta["generation"]

	if meta["namespace"] != "" {
		m["namespace"] = meta["namespace"]
	}

	return []map[string]interface{}{m}
}

func flattenVMSpec(obj map[string]interface{}) []map[string]interface{} {
	specglobal := obj["spec"].(map[string]interface{})
	template := specglobal["template"].(map[string]interface{})
	spec := template["spec"].(map[string]interface{})
	domain := spec["domain"].(map[string]interface{})
	devices := domain["devices"].(map[string]interface{})
	disks := devices["disks"].([]interface{})
	resources := domain["resources"].(map[string]interface{})
	requests := resources["requests"].(map[string]interface{})

	volumes := spec["volumes"].([]interface{})

	m := map[string]interface{}{
		"running": specglobal["running"].(bool),
		"memory":  requests["memory"].(string),
		"disks":   flattenVMDisksSpec(disks, volumes),
	}

	return []map[string]interface{}{m}
}

func flattenVMDisksSpec(disks []interface{}, volumes []interface{}) []map[string]interface{} {
	tdisks := make([]map[string]interface{}, len(disks))
	for i, v := range disks {
		tdisks[i] = map[string]interface{}{
			"name": v.(map[string]interface{})["name"].(string),
			"bus":  v.(map[string]interface{})["disk"].(map[string]interface{})["bus"].(string),
			"volume": map[string]string{
				"image": findVolumeImageByDiskName(v.(map[string]interface{})["name"].(string), volumes),
			},
		}
	}
	return tdisks
}

func findVolumeImageByDiskName(diskname string, volumes []interface{}) string {
	for _, v := range volumes {
		if v.(map[string]interface{})["name"] == diskname {
			return v.(map[string]interface{})["containerDisk"].(map[string]interface{})["image"].(string)
		}
	}

	return ""
}

func patchMetadata(keyPrefix, pathPrefix string, d *schema.ResourceData) PatchOperations {
	ops := make([]PatchOperation, 0, 0)
	if d.HasChange(keyPrefix + "annotations") {
		oldV, newV := d.GetChange(keyPrefix + "annotations")
		diffOps := diffStringMap(pathPrefix+"annotations", oldV.(map[string]interface{}), newV.(map[string]interface{}))
		ops = append(ops, diffOps...)
	}
	if d.HasChange(keyPrefix + "labels") {
		oldV, newV := d.GetChange(keyPrefix + "labels")
		diffOps := diffStringMap(pathPrefix+"labels", oldV.(map[string]interface{}), newV.(map[string]interface{}))
		ops = append(ops, diffOps...)
	}
	return ops
}

func patchVirtualMachineSpec(pathPrefix, prefix string, d *schema.ResourceData) (PatchOperations, error) {
	ops := make([]PatchOperation, 0)
	// TODO: implement

	return ops, nil
}
