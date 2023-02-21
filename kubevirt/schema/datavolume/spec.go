package datavolume

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/schema/k8s"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func dataVolumeSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"source": dataVolumeSourceSchema(),
		"pvc":    k8s.PersistentVolumeClaimSpecSchema(),
		"content_type": {
			Type:        schema.TypeString,
			Description: "ContentType options: \"kubevirt\", \"archive\".",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"kubevirt",
				"archive",
			}, false),
		},
	}
}

func DataVolumeSpecSchema() *schema.Schema {
	fields := dataVolumeSpecFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: fmt.Sprintf("DataVolumeSpec defines our specification for a DataVolume type"),
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func ExpandDataVolumeSpec(dataVolumeSpec []interface{}) (cdiv1.DataVolumeSpec, error) {
	result := cdiv1.DataVolumeSpec{}

	if len(dataVolumeSpec) == 0 || dataVolumeSpec[0] == nil {
		return result, nil
	}

	in := dataVolumeSpec[0].(map[string]interface{})

	result.Source = expandDataVolumeSource(in["source"].([]interface{}))
	p, err := k8s.ExpandPersistentVolumeClaimSpec(in["pvc"].([]interface{}))
	if err != nil {
		return result, err
	}
	result.PVC = p

	if v, ok := in["content_type"].(string); ok {
		result.ContentType = cdiv1.DataVolumeContentType(v)
	}

	return result, nil
}

func FlattenDataVolumeSpec(spec cdiv1.DataVolumeSpec) []interface{} {
	att := map[string]interface{}{
		"source":       flattenDataVolumeSource(spec.Source),
		"pvc":          k8s.FlattenPersistentVolumeClaimSpec(*spec.PVC),
		"content_type": string(spec.ContentType),
	}
	return []interface{}{att}
}
