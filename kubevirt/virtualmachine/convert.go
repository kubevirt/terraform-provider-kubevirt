package virtualmachine

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils"
	corev1 "k8s.io/api/core/v1"
	k8sv1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/client-go/api/v1"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
)

const (
	mainNetworkName = "main"
	defaultBus      = "virtio"
)

func updateResourceDataFromVirtualMachine(d *schema.ResourceData, vm *kubevirtapiv1.VirtualMachine) error {
	d.SetId(vm.GetName())
	if err := d.Set("name", vm.GetName()); err != nil {
		return err
	}
	if err := d.Set("namespace", vm.GetNamespace()); err != nil {
		return err
	}
	if err := d.Set("labels", vm.Labels); err != nil {
		return err
	}
	// ignitionSecretName
	// networkName
	// memory
	// cpu
	// accessMode
	// storageSize
	// storageClassName
	// pvcName
	return nil
}

func virtualMachineFromResourceData(d *schema.ResourceData) (*kubevirtapiv1.VirtualMachine, error) {
	// ResourceData Input
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	labels := d.Get("labels").(map[string]interface{})

	runAlways := kubevirtapiv1.RunStrategyAlways

	vmiTemplate, err := vmiTemplateFromResourceData(d)
	if err != nil {
		return nil, err
	}

	virtualMachine := kubevirtapiv1.VirtualMachine{
		Spec: kubevirtapiv1.VirtualMachineSpec{
			RunStrategy: &runAlways,
			DataVolumeTemplates: []cdiv1.DataVolume{
				*volumeTemplateFromResourceData(d),
			},
			Template: vmiTemplate,
		},
	}

	virtualMachine.APIVersion = kubevirtapiv1.GroupVersion.String()
	virtualMachine.Kind = "VirtualMachine"
	virtualMachine.ObjectMeta = metav1.ObjectMeta{
		Name:            name,
		Namespace:       namespace,
		Labels:          utils.ConvertMap(labels),
		OwnerReferences: nil,
	}

	return &virtualMachine, nil
}

func vmiTemplateFromResourceData(d *schema.ResourceData) (*kubevirtapiv1.VirtualMachineInstanceTemplateSpec, error) {
	// ResourceData Input
	name := d.Get("name").(string)
	ignitionSecretName := d.Get("ignition_secret_name").(string)
	serviceAccountName := d.Get("service_account_name").(string)
	networkName := d.Get("network_name").(string)
	memory := d.Get("memory").(string)
	cpu := d.Get("cpu").(int)
	antiAffinityTopologyKey := d.Get("anti_affinity_topology_key").(string)
	antiAffinityMatchLabels := d.Get("anti_affinity_match_labels").(map[string]interface{})

	template := &kubevirtapiv1.VirtualMachineInstanceTemplateSpec{}

	template.ObjectMeta = metav1.ObjectMeta{
		Labels: map[string]string{"kubevirt.io/vm": name, "name": name},
	}

	template.Spec = kubevirtapiv1.VirtualMachineInstanceSpec{}
	template.Spec.Volumes = []kubevirtapiv1.Volume{
		{
			Name: fmt.Sprintf("%s-datavolumedisk1", name),
			VolumeSource: kubevirtapiv1.VolumeSource{
				DataVolume: &kubevirtapiv1.DataVolumeSource{
					Name: fmt.Sprintf("%s-bootvolume", name),
				},
			},
		},
		{
			Name: fmt.Sprintf("%s-cloudinitdisk", name),
			VolumeSource: kubevirtapiv1.VolumeSource{
				CloudInitConfigDrive: &kubevirtapiv1.CloudInitConfigDriveSource{
					UserDataSecretRef: &corev1.LocalObjectReference{
						Name: ignitionSecretName,
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("%s-serviceaccountdisk", name),
			VolumeSource: kubevirtapiv1.VolumeSource{
				ServiceAccount: &kubevirtapiv1.ServiceAccountVolumeSource{
					ServiceAccountName: serviceAccountName,
				},
			},
		},
	}

	template.Spec.Networks = []kubevirtapiv1.Network{
		{
			Name: mainNetworkName,
			NetworkSource: kubevirtapiv1.NetworkSource{
				Multus: &kubevirtapiv1.MultusNetwork{
					NetworkName: networkName,
				},
			},
		},
	}

	template.Spec.Domain = kubevirtapiv1.DomainSpec{}

	template.Spec.Domain.Resources = kubevirtapiv1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: apiresource.MustParse(fmt.Sprint(memory)),
			corev1.ResourceCPU:    apiresource.MustParse(fmt.Sprint(cpu)),
		},
	}

	template.Spec.Domain.Devices = kubevirtapiv1.Devices{
		Disks: []kubevirtapiv1.Disk{
			{
				Name: fmt.Sprintf("%s-datavolumedisk1", name),
				DiskDevice: kubevirtapiv1.DiskDevice{
					Disk: &kubevirtapiv1.DiskTarget{
						Bus: defaultBus,
					},
				},
			},
			{
				Name: fmt.Sprintf("%s-cloudinitdisk", name),
				DiskDevice: kubevirtapiv1.DiskDevice{
					Disk: &kubevirtapiv1.DiskTarget{
						Bus: defaultBus,
					},
				},
			},
			{
				Name: fmt.Sprintf("%s-serviceaccountdisk", name),
				DiskDevice: kubevirtapiv1.DiskDevice{
					Disk: &kubevirtapiv1.DiskTarget{
						Bus: defaultBus,
					},
				},
			},
		},
		Interfaces: []kubevirtapiv1.Interface{
			{
				Name: mainNetworkName,
				InterfaceBindingMethod: kubevirtapiv1.InterfaceBindingMethod{
					Bridge: &kubevirtapiv1.InterfaceBridge{},
				},
			},
		},
	}

	if len(antiAffinityTopologyKey) > 0 {
		template.Spec.Affinity = &k8sv1.Affinity{
			PodAntiAffinity: &k8sv1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []k8sv1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: k8sv1.PodAffinityTerm{
							TopologyKey: antiAffinityTopologyKey,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: utils.ConvertMap(antiAffinityMatchLabels),
							},
						},
					},
				},
			},
		}
	}

	return template, nil
}

func volumeTemplateFromResourceData(d *schema.ResourceData) *cdiv1.DataVolume {
	// ResourceData Input
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	accessMode := d.Get("access_mode").(string)
	storageSize := d.Get("storage_size").(string)
	storageClassName := d.Get("storage_class_name").(string)
	pvcName := d.Get("pvc_name").(string)
	imageUrl := d.Get("image_url").(string)

	persistentVolumeClaimSpec := corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.PersistentVolumeAccessMode(accessMode),
		},
		// TODO: Where to get it?? - add as a list
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: apiresource.MustParse(storageSize),
			},
		},
	}
	if storageClassName != "" {
		persistentVolumeClaimSpec.StorageClassName = &storageClassName
	}

	dataVolumeSpec := cdiv1.DataVolumeSpec{
		Source: cdiv1.DataVolumeSource{},
		PVC:    &persistentVolumeClaimSpec,
	}
	if pvcName != "" {
		dataVolumeSpec.Source.PVC = &cdiv1.DataVolumeSourcePVC{
			Name:      pvcName,
			Namespace: namespace,
		}

	}
	if imageUrl != "" {
		dataVolumeSpec.Source.HTTP = &cdiv1.DataVolumeSourceHTTP{
			URL: imageUrl,
		}
	}
	return &cdiv1.DataVolume{
		TypeMeta: metav1.TypeMeta{APIVersion: cdiv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-bootvolume", name),
			Namespace: namespace,
		},
		Spec: dataVolumeSpec,
	}
}
