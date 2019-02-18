package kubevirt

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

// CloudInitDiskName is name of disk/volume which is used for cloud init created by terraform kubevirt provider
const CloudInitDiskName = "cloud_init_volume_terraform"

func expandVirtualMachineSpec(l []interface{}, metadata map[string]interface{}) (*map[string]interface{}, error) {
	obj := make(map[string]interface{})
	if len(l) == 0 || l[0] == nil {
		return &obj, nil
	}
	labels := metadata["labels"].(map[string]interface{})
	spec := l[0].(map[string]interface{})
	cloudInit := spec["cloud_init_no_cloud"].(map[string]interface{})
	memory := spec["memory"].([]interface{})[0].(map[string]interface{})
	obj = map[string]interface{}{
		"running": spec["running"].(bool),
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": labels,
			},
			"spec": map[string]interface{}{
				"domain": map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]string{
							"memory": memory["request"].(string),
						},
					},
					"devices": map[string]interface{}{
						"disks": expandVirtualMachineDisksSpec(spec["disks"].([]interface{}), &cloudInit),
					},
				},
				"volumes": expandVirtualMachineVolumesSpec(spec["disks"].([]interface{}), &cloudInit),
			},
		},
	}

	return &obj, nil
}

func expandVirtualMachineDisksSpec(d []interface{}, clInit *map[string]interface{}) []map[string]interface{} {
	// Initialize disks slice:
	disksSize := len(d)
	if len(*clInit) > 0 {
		disksSize++
	}
	disks := make([]map[string]interface{}, disksSize)

	// Add all user defined disks:
	for i, v := range d {
		disk := v.(map[string]interface{})
		diskspec := disk["disk"].(map[string]interface{})

		disks[i] = map[string]interface{}{
			"name": disk["name"].(string),
			"disk": map[string]string{
				"bus": diskspec["bus"].(string),
			},
		}
	}

	// Add cloud-init disk:
	if len(*clInit) > 0 {
		disks[disksSize-1] = map[string]interface{}{
			"name": "cloud_init_volume_terraform",
			"disk": map[string]string{
				"bus": "virtio",
			},
		}
	}
	return disks
}

func expandVirtualMachineVolumesSpec(d []interface{}, clInit *map[string]interface{}) []map[string]interface{} {
	// Initialize volumes slice:
	volumesSize := len(d)
	if len(*clInit) > 0 {
		volumesSize++
	}
	volumes := make([]map[string]interface{}, volumesSize)

	// Add all user defined volumes:
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

	// Add cloud-init volume:
	if len(*clInit) > 0 {
		volumes[volumesSize-1] = map[string]interface{}{
			"name":             "cloud_init_volume_terraform",
			"cloudInitNoCloud": *clInit,
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

func flattenVMSpec(specglobal map[string]interface{}) []map[string]interface{} {
	template := specglobal["template"].(map[string]interface{})
	spec := template["spec"].(map[string]interface{})
	domain := spec["domain"].(map[string]interface{})
	devices := domain["devices"].(map[string]interface{})
	disks := devices["disks"].([]interface{})
	//interfaces := devices["interfaces"].([]interface{})
	resources := domain["resources"].(map[string]interface{})
	requests := resources["requests"].(map[string]interface{})

	volumes := spec["volumes"].([]interface{})
	//networks := spec["networks"].([]interface{})

	m := map[string]interface{}{
		"running": specglobal["running"].(bool),
		"memory": [1]map[string]interface{}{
			{
				"request": requests["memory"].(string),
			},
		},
		"disks":               flattenVMDisksSpec(disks, volumes),
		"cloud_init_no_cloud": flattenCloudInitSpec(volumes),
		//"interfaces":          flattenVMInterfacesSpec(interfaces, networks),
	}

	return []map[string]interface{}{m}
}

func flattenCloudInitSpec(volumes []interface{}) map[string]interface{} {
	for _, v := range volumes {
		diskName := v.(map[string]interface{})["name"].(string)
		if diskName == CloudInitDiskName {
			return v.(map[string]interface{})["cloudInitNoCloud"].(map[string]interface{})
		}
	}

	return nil
}

func flattenVMDisksSpec(disks []interface{}, volumes []interface{}) []map[string]interface{} {
	tdisks := make([]map[string]interface{}, len(disks))
	for i, v := range disks {
		disk := v.(map[string]interface{})
		diskName := disk["name"].(string)
		if diskName == CloudInitDiskName {
			continue
		}
		tdisks[i] = map[string]interface{}{
			"name": diskName,
			"disk": map[string]string{
				"bus": disk["disk"].(map[string]interface{})["bus"].(string),
			},
			"volume": map[string]string{
				"image": findVolumeImageByDiskName(diskName, volumes),
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

	if d.HasChange(prefix + "running") {
		v := d.Get(prefix + "running").(bool)
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/running",
			Value: v,
		})
	}

	if d.HasChange(prefix + "memory") {
		v := d.Get(prefix + "memory").(string)
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/template/spec/domain/resources/requests/memory",
			Value: v,
		})
	}

	return ops, nil
}
