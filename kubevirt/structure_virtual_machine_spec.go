package kubevirt

import (
	"github.com/hashicorp/terraform/helper/schema"
)

// CloudInitDiskName is name of disk/volume which is used for cloud init created by terraform kubevirt provider
const CloudInitDiskName = "clinitterraform"

// DatavolumePrefix is prefix of volume created by terraform for datavolumes:
const DatavolumePrefix = "dvvolume"

func expandVirtualMachineSpec(d *schema.ResourceData) (*map[string]interface{}, error) {
	obj := make(map[string]interface{})
	labels := d.Get("labels").(map[string]interface{})
	annotations := d.Get("annotations").(map[string]interface{})
	cloudInit := d.Get("cloud_init").(map[string]interface{})
	name := d.Get("name").(string)
	image := d.Get("image").(interface{})
	memory := d.Get("memory").([]interface{})
	cpu := d.Get("cpu").([]interface{})
	interfaces := d.Get("interfaces").([]interface{})

	obj = map[string]interface{}{
		"running":             d.Get("running").(bool),
		"dataVolumeTemplates": expandImage(image, name),
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels":      labels,
				"annotations": annotations,
			},
			"spec": map[string]interface{}{
				"volumes": expandVolumes(name, image, cloudInit),
				"domain": map[string]interface{}{
					"devices":   expandDevices(name, image, cloudInit, interfaces),
					"cpu":       expandCPU(cpu),
					"resources": expandResources(memory, cpu),
				},
				"networks": expandNetworks(interfaces),
			},
		},
	}

	return &obj, nil
}

func expandVolumes(vmName string, image interface{}, cloudInit interface{}) []map[string]interface{} {
	volumes := make([]map[string]interface{}, 0)

	// Append qcow image volume:
	if len(image.(map[string]interface{})) > 0 {
		volumes = append(
			volumes,
			map[string]interface{}{
				"dataVolume": map[string]interface{}{
					"name": DatavolumePrefix + vmName,
				},
				"name": DatavolumePrefix + vmName,
			},
		)
	}

	// Append cloud-init volume:
	if len(cloudInit.(map[string]interface{})) > 0 {
		volumes = append(
			volumes,
			map[string]interface{}{
				"cloudInitNoCloud": cloudInit.(map[string]interface{}),
				"name":             CloudInitDiskName,
			},
		)
	}

	return volumes
}

func expandDevices(vmName string, image interface{}, cloudInit interface{}, interfaces []interface{}) map[string]interface{} {
	devices := make(map[string]interface{})
	disks := make([]map[string]interface{}, 0)

	// Append qcow image disk:
	if len(image.(map[string]interface{})) > 0 {
		disks = append(
			disks,
			map[string]interface{}{
				"disk": map[string]interface{}{
					"bus": "virtio",
				},
				"name": DatavolumePrefix + vmName,
			},
		)
	}

	// Append cloud-init disk:
	if len(cloudInit.(map[string]interface{})) > 0 {
		disks = append(
			disks,
			map[string]interface{}{
				"name": CloudInitDiskName,
				"disk": map[string]string{
					"bus": "virtio",
				},
			},
		)
	}

	if len(disks) > 0 {
		devices["disks"] = disks
	}

	if len(interfaces) > 0 {
		devices["interfaces"] = expandInterfaces(interfaces)
	}

	return devices
}

func expandInterfaces(interfacesIn []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(interfacesIn))

	for i, nic := range interfacesIn {
		nic := nic.(map[string]interface{})
		result[i] = make(map[string]interface{})
		nicType := nic["type"].(string)
		result[i][nicType] = map[string]string{}
		result[i]["name"] = nic["name"]
		result[i]["ports"] = nic["ports"]
		result[i]["pciAddress"] = nic["pci_address"]
		result[i]["model"] = nic["model"]
		result[i]["macAddress"] = nic["mac_address"]
		if nic["boot_order"] != 0 {
			result[i]["bootOrder"] = nic["boot_order"]
		}
	}

	return result
}

func expandNetworks(interfacesIn []interface{}) []map[string]interface{} {
	if len(interfacesIn) == 0 {
		return nil
	}
	result := make([]map[string]interface{}, len(interfacesIn))

	for i, nic := range interfacesIn {
		nic := nic.(map[string]interface{})
		result[i] = make(map[string]interface{})
		nicNetwork := nic["network"].(string)
		result[i]["name"] = nic["name"]
		result[i][nicNetwork] = map[string]string{}
	}

	return result
}

func expandCPU(cpuIn []interface{}) map[string]interface{} {
	if len(cpuIn) == 0 {
		return nil
	}
	cpu := make(map[string]interface{})
	if len(cpuIn) > 0 {
		cpu = cpuIn[0].(map[string]interface{})
	}
	result := make(map[string]interface{})
	result["sockets"] = cpu["sockets"]
	result["model"] = cpu["model"]
	result["dedicatedCpuPlacement"] = cpu["dedicated"]
	result["threads"] = cpu["threads"]
	result["cores"] = cpu["cores"]

	return result
}

func expandResources(memoryIn []interface{}, cpuIn []interface{}) map[string]map[string]string {
	if len(memoryIn) == 0 && len(cpuIn) == 0 {
		return nil
	}
	memory := make(map[string]interface{})
	if len(memoryIn) > 0 {
		memory = memoryIn[0].(map[string]interface{})
	}
	cpu := make(map[string]interface{})
	if len(cpuIn) > 0 {
		cpu = cpuIn[0].(map[string]interface{})
	}
	resources := map[string]map[string]string{
		"requests": map[string]string{},
		"limits":   map[string]string{},
	}

	// Process memory limit:
	mLimit := memory["limits"]
	if mLimit != "" {
		resources["limits"]["memory"] = mLimit.(string)
	}

	// Process requests memory:
	mReq := memory["requests"]
	if mReq != "" {
		resources["requests"]["memory"] = mReq.(string)
	}

	// Process requests cpu:
	cpuReq := cpu["requests"]
	if cpuReq != "" {
		resources["requests"]["cpu"] = cpuReq.(string)
	}

	// Process cpu limit:
	cpuLimit := cpu["limits"]
	if cpuLimit != "" {
		resources["limits"]["cpu"] = cpuLimit.(string)
	}

	return resources
}

func expandImage(img interface{}, name string) []map[string]interface{} {
	if img == nil {
		return nil
	}
	image := img.(map[string]interface{})
	accessModes, ok := image["accessmodes"].([]interface{})
	if !ok {
		accessModes = []interface{}{"ReadWriteOnce"}
	}
	storage, ok := image["storage"].(string)
	if !ok {
		storage = "1Gi"
	}

	return []map[string]interface{}{
		map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": DatavolumePrefix + name,
			},
			"spec": map[string]interface{}{
				"pvc": map[string]interface{}{
					"accessModes": accessModes,
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": storage,
						},
					},
				},
				"source": map[string]interface{}{
					"http": map[string]interface{}{
						"url": image["url"].(string),
					},
				},
			},
		},
	}
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

func flattenVMInterfacesSpec(interfaces, networks []interface{}) []map[string]interface{} {
	flatten := make([]map[string]interface{}, len(interfaces))
	for index, iface := range interfaces {
		iface := iface.(map[string]interface{})
		flatten[index] = make(map[string]interface{})
		flatten[index]["name"] = iface["name"]
		flatten[index]["type"] = iface["type"]
		flatten[index]["network"] = iface["network"]
	}

	return flatten
}
