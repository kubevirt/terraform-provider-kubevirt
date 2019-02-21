package kubevirt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

// CloudInitDiskName is name of disk/volume which is used for cloud init created by terraform kubevirt provider
const CloudInitDiskName = "clinitterraform"

// DatavolumePrefix is prefix of volume created by terraform for datavolumes:
const DatavolumePrefix = "dvvolume"

func expandVirtualMachineSpec(l []interface{}, metadata map[string]interface{}) (*map[string]interface{}, error) {
	obj := make(map[string]interface{})
	if len(l) == 0 || l[0] == nil {
		return &obj, nil
	}
	labels := metadata["labels"].(map[string]interface{})
	spec := l[0].(map[string]interface{})
	cloudInit := spec["cloud_init_no_cloud"].(map[string]interface{})
	dv := spec["datavolumes"].([]interface{})
	memory := spec["memory"].([]interface{})[0].(map[string]interface{})

	obj = map[string]interface{}{
		"running":             spec["running"].(bool),
		"dataVolumeTemplates": expandDatavolumes(dv),
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
						"disks": expandVirtualMachineDisksSpec(spec["disks"].([]interface{}), &cloudInit, dv),
					},
				},
				"volumes": expandVirtualMachineVolumesSpec(spec["disks"].([]interface{}), &cloudInit, dv),
			},
		},
	}

	return &obj, nil
}

func expandDatavolumes(dvs []interface{}) []map[string]interface{} {
	dvsspec := make([]map[string]interface{}, len(dvs))
	for i, dv := range dvs {
		dv := dv.(map[string]interface{})
		source := dv["source"].([]interface{})[0].(map[string]interface{})
		pvc := dv["pvc"].([]interface{})[0].(map[string]interface{})

		dvsspec[i] = make(map[string]interface{})
		dvsspec[i]["metadata"] = map[string]interface{}{
			"name": dv["name"].(string),
		}
		dvsspec[i]["spec"] = map[string]interface{}{
			"pvc": map[string]interface{}{
				"accessModes": pvc["accessmodes"].([]interface{}),
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"storage": pvc["storage"].(string),
					},
				},
			},
			"source": map[string]interface{}{
				"http": source["http"].(map[string]interface{}),
			},
		}
	}

	return dvsspec
}

func expandVirtualMachineDisksSpec(d []interface{}, clInit *map[string]interface{}, dvs []interface{}) []map[string]interface{} {
	// Initialize disks slice:
	disksSize := len(d)
	if len(*clInit) > 0 {
		disksSize++
	}
	if len(dvs) > 0 {
		disksSize += len(dvs)
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
		disks[disksSize-1-len(dvs)] = map[string]interface{}{
			"name": CloudInitDiskName,
			"disk": map[string]string{
				"bus": "virtio",
			},
		}
	}
	// Add datavolume disk:
	if len(dvs) > 0 {
		for i, dv := range dvs {
			dv := dv.(map[string]interface{})
			disks[disksSize-1-i] = map[string]interface{}{
				"name": fmt.Sprintf("%s%s", DatavolumePrefix, dv["name"].(string)),
				"disk": map[string]string{
					"bus": "virtio",
				},
			}
		}
	}
	return disks
}

func expandVirtualMachineVolumesSpec(d []interface{}, clInit *map[string]interface{}, dvs []interface{}) []map[string]interface{} {
	// Initialize volumes slice:
	volumesSize := len(d)
	if len(*clInit) > 0 {
		volumesSize++
	}
	if len(dvs) > 0 {
		volumesSize += len(dvs)
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
		volumes[volumesSize-1-len(dvs)] = map[string]interface{}{
			"name":             CloudInitDiskName,
			"cloudInitNoCloud": *clInit,
		}
	}
	// Add datavolume volume:
	if len(dvs) > 0 {
		for i, dv := range dvs {
			dv := dv.(map[string]interface{})
			volumes[volumesSize-1-i] = map[string]interface{}{
				"name": fmt.Sprintf("%s%s", DatavolumePrefix, dv["name"].(string)),
				"dataVolume": map[string]string{
					"name": dv["name"].(string),
				},
			}
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
	dvs, _ := specglobal["dataVolumeTemplates"].([]interface{})
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
		"datavolumes":         flattenDatavolumesSpec(dvs),
		//"interfaces":          flattenVMInterfacesSpec(interfaces, networks),
	}

	return []map[string]interface{}{m}
}

func flattenDatavolumesSpec(dvsspec []interface{}) []map[string]interface{} {
	dvs := make([]map[string]interface{}, len(dvsspec))
	for i, dv := range dvsspec {
		dv := dv.(map[string]interface{})
		metadata := dv["metadata"].(map[string]interface{})
		spec := dv["spec"].(map[string]interface{})
		pvc := spec["pvc"].(map[string]interface{})
		source := spec["source"].(map[string]interface{})

		dvs[i] = make(map[string]interface{})
		dvs[i]["name"] = metadata["name"].(string)
		dvs[i]["source"] = [1]map[string]interface{}{{
			"http": source["http"].(map[string]interface{}),
		}}
		dvs[i]["pvc"] = [1]map[string]interface{}{{
			"accessmodes": pvc["accessModes"].([]interface{}),
			"storage":     pvc["resources"].(map[string]interface{})["requests"].(map[string]interface{})["storage"].(string),
		}}
	}

	return dvs
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
		if strings.HasPrefix(diskName, DatavolumePrefix) {
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
		memory := d.Get(prefix + "memory").([]interface{})[0].(map[string]interface{})
		v := memory["request"].(string)
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/template/spec/domain/resources/requests/memory",
			Value: v,
		})
	}

	return ops, nil
}
