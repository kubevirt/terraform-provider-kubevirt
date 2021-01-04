package flatten_utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	test_entities "github.com/kubevirt/terraform-provider-kubevirt/kubevirt/test_utils/entities"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"

	kubevirtapiv1 "kubevirt.io/client-go/api/v1"
)

func GetBaseInputForDataVolume() cdiv1.DataVolume {
	return cdiv1.DataVolume{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "generate_name",
			Name:         "test-vm-bootvolume",
			Namespace:    "tenantcluster",
		},
		Spec: cdiv1.DataVolumeSpec{
			Source: cdiv1.DataVolumeSource{
				HTTP: &cdiv1.DataVolumeSourceHTTP{
					URL:           "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2",
					SecretRef:     "secret_ref",
					CertConfigMap: "cert_config_map",
				},
				PVC: &cdiv1.DataVolumeSourcePVC{
					Namespace: "namespace",
					Name:      "name",
				},
			},
			PVC: &k8sv1.PersistentVolumeClaimSpec{
				AccessModes: []k8sv1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: k8sv1.ResourceRequirements{
					Requests: k8sv1.ResourceList{
						"storage": (func() resource.Quantity { res, _ := resource.ParseQuantity("10Gi"); return res })(),
					},
					Limits: k8sv1.ResourceList{
						"storage": (func() resource.Quantity { res, _ := resource.ParseQuantity("20Gi"); return res })(),
					},
				},
				Selector:         test_entities.LabelSelectorAPI,
				VolumeName:       "volume_name",
				StorageClassName: (func() *string { str := "standard"; return &str })(),
			},
			ContentType: cdiv1.DataVolumeContentType("content_type"),
		},
	}
}

func GetBaseOutputForDataVolume() interface{} {
	return map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"annotations":      interface{}(map[string]interface{}(nil)),
				"labels":           interface{}(map[string]interface{}(nil)),
				"name":             "test-vm-bootvolume",
				"resource_version": interface{}(""),
				"self_link":        interface{}(""),
				"uid":              interface{}(""),
				"generation":       interface{}(int64(0)),
				"namespace":        "tenantcluster",
				"generate_name":    "generate_name",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"pvc": []interface{}{
					map[string]interface{}{
						"access_modes": (func() *schema.Set {
							out := []interface{}{
								"ReadWriteOnce",
							}
							return schema.NewSet(schema.HashString, out)
						})(),
						"resources": []interface{}{
							map[string]interface{}{
								"requests": map[string]interface{}{
									"storage": "10Gi",
								},
								"limits": map[string]interface{}{
									"storage": "20Gi",
								},
							},
						},
						"selector":           test_entities.LabelSelectorTerraform,
						"volume_name":        "volume_name",
						"storage_class_name": "standard",
					},
				},
				"source": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								"url":             "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2",
								"secret_ref":      "secret_ref",
								"cert_config_map": "cert_config_map",
							},
						},
						"pvc": []interface{}{
							map[string]interface{}{
								"namespace": "namespace",
								"name":      "name",
							},
						},
					},
				},
				"content_type": "content_type",
			},
		},
		"status": []interface{}{
			map[string]interface{}{
				"phase":    "",
				"progress": "",
			},
		},
	}
}

func GetBaseInputForVirtualMachine() kubevirtapiv1.VirtualMachineSpec {
	return kubevirtapiv1.VirtualMachineSpec{
		RunStrategy: (func() *kubevirtapiv1.VirtualMachineRunStrategy {
			strategy := kubevirtapiv1.VirtualMachineRunStrategy("Always")
			return &strategy
		})(),
		DataVolumeTemplates: []cdiv1.DataVolume{
			GetBaseInputForDataVolume(),
		},
		Template: &kubevirtapiv1.VirtualMachineInstanceTemplateSpec{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{
					"annotation_key": "annotation_value",
				},
				Labels: map[string]string{
					"kubevirt.io/vm": "test-vm",
				},
				GenerateName: "generate_name",
				Name:         "name",
				Namespace:    "namespace",
			},
			Spec: kubevirtapiv1.VirtualMachineInstanceSpec{
				PriorityClassName: "priority_class_name",
				Volumes: []kubevirtapiv1.Volume{
					{
						Name: "test-vm-datavolumedisk1",
						VolumeSource: kubevirtapiv1.VolumeSource{
							DataVolume: &kubevirtapiv1.DataVolumeSource{
								Name: "test-vm-bootvolume",
							},
							CloudInitConfigDrive: &kubevirtapiv1.CloudInitConfigDriveSource{
								UserDataSecretRef: &k8sv1.LocalObjectReference{
									Name: "name",
								},
								UserDataBase64: "user_data_base64",
								UserData:       "user_data",
								NetworkDataSecretRef: &k8sv1.LocalObjectReference{
									Name: "name",
								},
								NetworkDataBase64: "network_data_base64",
								NetworkData:       "network_data",
							},
							ServiceAccount: &kubevirtapiv1.ServiceAccountVolumeSource{
								ServiceAccountName: "service_account_name",
							},
						},
					},
				},
				Domain: kubevirtapiv1.DomainSpec{
					Resources: kubevirtapiv1.ResourceRequirements{
						Requests: k8sv1.ResourceList{
							"memory": (func() resource.Quantity { res, _ := resource.ParseQuantity("10G"); return res })(),
							"cpu":    *resource.NewQuantity(int64(4), resource.DecimalExponent),
						},
						Limits: k8sv1.ResourceList{
							"memory": (func() resource.Quantity { res, _ := resource.ParseQuantity("20G"); return res })(),
							"cpu":    *resource.NewQuantity(int64(8), resource.DecimalExponent),
						},
						OvercommitGuestOverhead: true,
					},
					Devices: kubevirtapiv1.Devices{
						Disks: []kubevirtapiv1.Disk{
							{
								Name:   "test-vm-datavolumedisk1",
								Serial: "serial",
								DiskDevice: kubevirtapiv1.DiskDevice{
									Disk: &kubevirtapiv1.DiskTarget{
										Bus:        "virtio",
										ReadOnly:   true,
										PciAddress: "pci_address",
									},
								},
							},
						},
						Interfaces: []kubevirtapiv1.Interface{
							{
								Name: "main",
								InterfaceBindingMethod: kubevirtapiv1.InterfaceBindingMethod{
									Bridge: &kubevirtapiv1.InterfaceBridge{},
								},
							},
						},
					},
				},
				NodeSelector: map[string]string{
					"node_selector_key": "node_selector_value",
				},
				Hostname:      "hostname",
				Subdomain:     "subdomain",
				SchedulerName: "scheduler_name",
				Tolerations: []k8sv1.Toleration{
					{
						Effect:            k8sv1.TaintEffect("effect"),
						Key:               "key",
						Operator:          k8sv1.TolerationOperator("operator"),
						TolerationSeconds: utils.PtrToInt64(int64(60)),
						Value:             "value",
					},
				},
				EvictionStrategy: (func() *kubevirtapiv1.EvictionStrategy {
					retval := kubevirtapiv1.EvictionStrategy("eviction_strategy")
					return &retval
				})(),
				TerminationGracePeriodSeconds: utils.PtrToInt64(int64(120)),
				Networks: []kubevirtapiv1.Network{
					{
						Name: "main",
						NetworkSource: kubevirtapiv1.NetworkSource{
							Pod: &kubevirtapiv1.PodNetwork{
								VMNetworkCIDR: "vm_network_cidr",
							},
							Multus: &kubevirtapiv1.MultusNetwork{
								NetworkName: "tenantcluster",
							},
						},
					},
				},
				DNSPolicy: k8sv1.DNSPolicy("dns_policy"),
				DNSConfig: &k8sv1.PodDNSConfig{
					Options: []k8sv1.PodDNSConfigOption{
						{
							Name:  "name",
							Value: (func() *string { retval := "value"; return &retval })(),
						},
					},
				},
				Affinity: &k8sv1.Affinity{
					NodeAffinity: &k8sv1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &k8sv1.NodeSelector{
							NodeSelectorTerms: test_entities.NodeSelectorTermAPI,
						},
						PreferredDuringSchedulingIgnoredDuringExecution: []k8sv1.PreferredSchedulingTerm{
							{
								Weight: int32(10),
								Preference: k8sv1.NodeSelectorTerm{
									MatchExpressions: test_entities.MatchExpressionAPI,
									MatchFields:      test_entities.MatchFieldsAPI,
								},
							},
						},
					},
					PodAffinity: &k8sv1.PodAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: test_entities.PodPreferredDuringSchedulingIgnoredDuringExecutionAPI,
						RequiredDuringSchedulingIgnoredDuringExecution:  test_entities.PodRequiredDuringSchedulingIgnoredDuringExecutionAPI,
					},
					PodAntiAffinity: &k8sv1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: test_entities.PodPreferredDuringSchedulingIgnoredDuringExecutionAPI,
						RequiredDuringSchedulingIgnoredDuringExecution:  test_entities.PodRequiredDuringSchedulingIgnoredDuringExecutionAPI,
					},
				},
			},
		},
	}
}

func GetBaseOutputForVirtualMachine() interface{} {
	return map[string]interface{}{
		"data_volume_templates": []interface{}{
			GetBaseOutputForDataVolume(),
		},
		"run_strategy": "Always",
		"template": []interface{}{
			map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"annotations": map[string]interface{}{
							"annotation_key": "annotation_value",
						},
						"labels": map[string]interface{}{
							"kubevirt.io/vm": "test-vm",
						},
						"generate_name":    "generate_name",
						"name":             "name",
						"namespace":        "namespace",
						"resource_version": interface{}(""),
						"self_link":        interface{}(""),
						"uid":              interface{}(""),
						"generation":       interface{}(int64(0)),
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"node_selector": map[string]interface{}{
							"node_selector_key": "node_selector_value",
						},
						"scheduler_name": "scheduler_name",
						"tolerations": []interface{}{
							map[string]interface{}{
								"effect":             "effect",
								"key":                "key",
								"operator":           "operator",
								"toleration_seconds": "60",
								"value":              "value",
							},
						},
						"dns_policy":          "dns_policy",
						"priority_class_name": "priority_class_name",
						"hostname":            "hostname",
						"subdomain":           "subdomain",
						"pod_dns_config": []interface{}{
							map[string]interface{}{
								"option": []interface{}{
									map[string]interface{}{
										"name":  "name",
										"value": "value",
									},
								},
							},
						},
						"affinity": []interface{}{
							map[string]interface{}{
								"node_affinity": []interface{}{
									map[string]interface{}{
										"required_during_scheduling_ignored_during_execution":  test_entities.NodeRequiredDuringSchedulingIgnoredDuringExecution,
										"preferred_during_scheduling_ignored_during_execution": test_entities.NodePreferredDuringSchedulingIgnoredDuringExecution,
									},
								},
								"pod_affinity": []interface{}{
									map[string]interface{}{
										"preferred_during_scheduling_ignored_during_execution": test_entities.PodPreferredDuringSchedulingIgnoredDuringExecutionTerraform,
										"required_during_scheduling_ignored_during_execution":  test_entities.PodRequiredDuringSchedulingIgnoredDuringExecutionTerraform,
									},
								},
								"pod_anti_affinity": []interface{}{
									map[string]interface{}{
										"preferred_during_scheduling_ignored_during_execution": test_entities.PodPreferredDuringSchedulingIgnoredDuringExecutionTerraform,
										"required_during_scheduling_ignored_during_execution":  test_entities.PodRequiredDuringSchedulingIgnoredDuringExecutionTerraform,
									},
								},
							},
						},
						"domain": []interface{}{
							map[string]interface{}{
								"devices": []interface{}{
									map[string]interface{}{
										"disk": []interface{}{
											map[string]interface{}{
												"disk_device": []interface{}{
													map[string]interface{}{
														"disk": []interface{}{
															map[string]interface{}{
																"bus":         "virtio",
																"read_only":   true,
																"pci_address": "pci_address",
															},
														},
													},
												},
												"name":   "test-vm-datavolumedisk1",
												"serial": "serial",
											},
										},
										"interface": []interface{}{
											map[string]interface{}{
												"interface_binding_method": "InterfaceBridge",
												"name":                     "main",
											},
										},
									},
								},
								"resources": []interface{}{
									map[string]interface{}{
										"requests": map[string]interface{}{
											"cpu":    "4",
											"memory": "10G",
										},
										"limits": map[string]interface{}{
											"cpu":    "8",
											"memory": "20G",
										},
										"over_commit_guest_overhead": true,
									},
								},
							},
						},
						"eviction_strategy":                "eviction_strategy",
						"termination_grace_period_seconds": int64(120),
						"volume": []interface{}{
							map[string]interface{}{
								"name": "test-vm-datavolumedisk1",
								"volume_source": []interface{}{
									map[string]interface{}{
										"data_volume": []interface{}{
											map[string]interface{}{
												"name": "test-vm-bootvolume",
											},
										},
										"cloud_init_config_drive": []interface{}{
											map[string]interface{}{
												"user_data_secret_ref": []interface{}{
													map[string]interface{}{
														"name": "name",
													},
												},
												"user_data_base64": "user_data_base64",
												"user_data":        "user_data",
												"network_data_secret_ref": []interface{}{
													map[string]interface{}{
														"name": "name",
													},
												},
												"network_data_base64": "network_data_base64",
												"network_data":        "network_data",
											},
										},
										"service_account": []interface{}{
											map[string]interface{}{
												"service_account_name": "service_account_name",
											},
										},
									},
								},
							},
						},
						"network": []interface{}{
							map[string]interface{}{
								"name": "main",
								"network_source": []interface{}{
									map[string]interface{}{
										"pod": []interface{}{
											map[string]interface{}{
												"vm_network_cidr": "vm_network_cidr",
											},
										},
										"multus": []interface{}{
											map[string]interface{}{
												"network_name": "tenantcluster",
												"default":      false,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
