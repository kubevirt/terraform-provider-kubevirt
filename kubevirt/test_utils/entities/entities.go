package entities

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt/utils"
	k8sv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var LabelSelectorTerraform = []interface{}{
	map[string]interface{}{
		"match_labels": map[string]interface{}{
			"anti-affinity-key": "anti-affinity-val",
		},
	},
}

var LabelSelectorAPI = &v1.LabelSelector{
	MatchLabels: map[string]string{
		"anti-affinity-key": "anti-affinity-val",
	},
}

var MatchExpressionTerraform = []interface{}{
	map[string]interface{}{
		"key":      "key",
		"operator": "operator",
		"values":   utils.NewStringSet(schema.HashString, []string{"value1", "value2"}),
	},
}

var MatchFieldsTerraform = []interface{}{
	map[string]interface{}{
		"key":      "key",
		"operator": "operator",
		"values":   utils.NewStringSet(schema.HashString, []string{"value1", "value2"}),
	},
}

var NodeSelectorTermTerraform = []interface{}{
	map[string]interface{}{
		"match_expressions": MatchExpressionTerraform,
		"match_fields":      MatchFieldsTerraform,
	},
}

var MatchExpressionAPI = []k8sv1.NodeSelectorRequirement{
	{
		Key:      "key",
		Operator: k8sv1.NodeSelectorOperator("operator"),
		Values:   []string{"value1", "value2"},
	},
}

var MatchFieldsAPI = []k8sv1.NodeSelectorRequirement{
	{
		Key:      "key",
		Operator: k8sv1.NodeSelectorOperator("operator"),
		Values:   []string{"value1", "value2"},
	},
}

var NodeSelectorTermAPI = []k8sv1.NodeSelectorTerm{
	{
		MatchExpressions: MatchExpressionAPI,
		MatchFields:      MatchFieldsAPI,
	},
}

var NodePreferredDuringSchedulingIgnoredDuringExecution = []interface{}{
	map[string]interface{}{
		"weight":     10,
		"preference": NodeSelectorTermTerraform,
	},
}

var NodeRequiredDuringSchedulingIgnoredDuringExecution = []interface{}{
	map[string]interface{}{
		"node_selector_term": NodeSelectorTermTerraform,
	},
}

var PodPreferredDuringSchedulingIgnoredDuringExecutionAPI = []k8sv1.WeightedPodAffinityTerm{
	{
		Weight: 100,
		PodAffinityTerm: k8sv1.PodAffinityTerm{
			LabelSelector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"anti-affinity-key": "anti-affinity-val",
				},
			},
			TopologyKey: "kubernetes.io/hostname",
			Namespaces:  []string{"namespace1"},
		},
	},
}

var PodPreferredDuringSchedulingIgnoredDuringExecutionTerraform = []interface{}{
	map[string]interface{}{
		"weight": 100,
		"pod_affinity_term": []interface{}{
			map[string]interface{}{
				"label_selector": LabelSelectorTerraform,
				"topology_key":   "kubernetes.io/hostname",
				"namespaces":     utils.NewStringSet(schema.HashString, []string{"namespace1"}),
			},
		},
	},
}

var PodRequiredDuringSchedulingIgnoredDuringExecutionAPI = []k8sv1.PodAffinityTerm{
	{
		LabelSelector: LabelSelectorAPI,
		TopologyKey:   "kubernetes.io/hostname",
		Namespaces:    []string{"namespace1"},
	},
}

var PodRequiredDuringSchedulingIgnoredDuringExecutionTerraform = []interface{}{
	map[string]interface{}{
		"label_selector": LabelSelectorTerraform,
		"topology_key":   "kubernetes.io/hostname",
		"namespaces":     utils.NewStringSet(schema.HashString, []string{"namespace1"}),
	},
}