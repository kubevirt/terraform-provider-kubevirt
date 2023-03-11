package virtualmachineinstance

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func probeFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"initial_delay_seconds": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "InitialDelaySeconds is the time to wait before starting to probe after pod startup. This is only applicable for Readiness and Liveness probes, ignored for HTTP probes.",
			Default:     600,
		},
		"period_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "PeriodSeconds specifies how often (in seconds between probes) to perform the probe.",
			Default:      60,
			ValidateFunc: validation.IntBetween(1, 300),
			// TODO: https://github.com/kubevirt/client-go/issues/8
			// Deprecated: true,
		},
		"failure_threshold": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "FailureThreshold specifies the number of consecutive failures before a pod is marked failed. It is only applicable for Readiness and Liveness probes, ignored for HTTP probes.",
			Default:      3,
			ValidateFunc: validation.IntBetween(1, 10),
			// TODO: https://github.com/kubevirt/client-go/issues/8
			// Deprecated: true,
			//Removed:      true,
		},
		"success_threshold": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "SuccessThreshold specifies the number of consecutive successes before the probe is considered successful. It is only applicable for Readiness and Liveness probes, ignoredConfigMode: Readiness and Liveness probes, ignored for HTTP probes.",
			Default:      1,
			ValidateFunc: validation.IntBetween(1, 10),
		},

		"tcp_socket": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "TCP socket probe",
			Elem: &schema.Resource{

				Schema: map[string]*schema.Schema{
					"port": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Port number",
					},
					"host": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Host name",
					},
				},
			},
		},

		"http_get": {
			Type: schema.TypeList,
			//Optional:    true,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{

				Schema: map[string]*schema.Schema{
					"path": {
						Type:        schema.TypeString,
						Description: "Path to access on the HTTP server.",
						Optional:    true,
						Default:     "/",
					},
					"port": {
						Type:        schema.TypeInt,
						Description: "Number or name of the port to access on the container. Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.",
						Optional:    true,
						Default:     80,
					},
					"host": {
						Type:        schema.TypeString,
						Description: "Host name to connect to, defaults to the pod IP.",
						Optional:    true,
						Default:     "",
					},
					"scheme": {
						Type:        schema.TypeString,
						Description: "Scheme to use for connecting to the host. Defaults to HTTP.",
						Optional:    true,
						Default:     "HTTP",
						ValidateFunc: validation.StringInSlice([]string{
							"HTTP",
							"HTTPS",
						}, false),
					},
					"http_headers": {
						Type:        schema.TypeList,
						Description: "HTTPHeader describes a custom header to be used in HTTP probes",
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Description: "The header field name",
									Optional:    true,
									Default:     "",
								},
								"value": {
									Type:        schema.TypeString,
									Description: "The header field value",
									Optional:    true,
									Default:     "",
								},
							},
						},
					},
				},
			},
		},
	}
}

func probeSchema() *schema.Schema {
	fields := probeFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: fmt.Sprintf("Specification of the desired behavior of the VirtualMachineInstance on the host."),
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandProbe(probe []interface{}) *kubevirtapiv1.Probe {
	if len(probe) == 0 || probe[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.Probe{}

	probeMap := probe[0].(map[string]interface{})
	if v, ok := probeMap["initial_delay_seconds"]; ok {
		result.InitialDelaySeconds = int32(v.(int))
	}
	if v, ok := probeMap["period_seconds"]; ok {
		result.PeriodSeconds = int32(v.(int))
	}
	if v, ok := probeMap["failure_threshold"]; ok {
		result.FailureThreshold = int32(v.(int))
	}
	if v, ok := probeMap["success_threshold"]; ok {
		result.SuccessThreshold = int32(v.(int))
	}
	if v, ok := probeMap["timeout_seconds"]; ok {
		result.TimeoutSeconds = int32(v.(int))
	}

	if v, ok := probeMap["http_get"]; ok {
		result.HTTPGet = expandHTTPGet(v.([]interface{}))
	}
	if v, ok := probeMap["tcp_socket"]; ok {
		result.TCPSocket = expandTCPSocket(v.([]interface{}))
	}

	return result
}

func expandHTTPGet(httpGet []interface{}) *k8sv1.HTTPGetAction {
	if len(httpGet) == 0 || httpGet[0] == nil {
		return nil
	}

	result := &k8sv1.HTTPGetAction{}

	httpGetMap := httpGet
	if v, ok := httpGetMap[0].(map[string]interface{}); ok {
		if v, ok := v["path"]; ok {
			result.Path = v.(string)
		}
		if v, ok := v["port"]; ok {
			switch v.(type) {
			case string:
				result.Port = intstr.Parse(v.(string))
			case int:
				result.Port = intstr.FromInt(v.(int))
			}
		}
		if v, ok := v["scheme"]; ok {
			result.Scheme = k8sv1.URIScheme(v.(string))
			if result.Scheme == "https" {
				result.Port = intstr.FromInt(443)
			}
		}
		if v, ok := v["http_header"]; ok {
			result.HTTPHeaders = expandHTTPHeader(v.([]interface{}))
		}
	}

	return result
}

func expandHTTPHeader(httpHeader []interface{}) []k8sv1.HTTPHeader {
	if len(httpHeader) == 0 || httpHeader[0] == nil {
		return nil
	}

	result := make([]k8sv1.HTTPHeader, 0, len(httpHeader))

	for _, h := range httpHeader {
		header := h.(map[string]interface{})
		result = append(result, k8sv1.HTTPHeader{
			Name:  header["name"].(string),
			Value: header["value"].(string),
		})
	}

	return result
}

func expandTCPSocket(tcpSocket []interface{}) *k8sv1.TCPSocketAction {
	if len(tcpSocket) == 0 || tcpSocket[0] == nil {
		return nil
	}

	result := &k8sv1.TCPSocketAction{}

	tcpSocketMap := tcpSocket
	if v, ok := tcpSocketMap[0].(map[string]interface{}); ok {
		if v, ok := v["host"]; ok {
			result.Host = v.(string)
		}

		if v, ok := v["port"]; ok {
			// result.Port = intstr.FromString(v.(string))
			switch v.(type) {
			case string:
				result.Port = intstr.FromString(v.(string))
			case int:
				result.Port = intstr.FromInt(v.(int))
			}
		}
	}

	return result
}

func flattenProbe(in kubevirtapiv1.Probe) []interface{} {
	att := make(map[string]interface{})

	if in.FailureThreshold != 0 {
		att["failure_threshold"] = int(in.FailureThreshold)
	}
	if in.SuccessThreshold != 0 {
		att["success_threshold"] = int(in.SuccessThreshold)
	}
	if in.PeriodSeconds != 0 {
		att["period_seconds"] = int(in.PeriodSeconds)
	}
	if in.TimeoutSeconds != 0 {
		att["timeout_seconds"] = int(in.TimeoutSeconds)
	}
	if in.HTTPGet != nil {
		att["http_get"] = flattenHTTPGet(in.HTTPGet)
	}
	if in.TCPSocket != nil {
		att["tcp_socket"] = flattenTCPSocket(in.TCPSocket)
	}
	if in.InitialDelaySeconds != 0 {
		att["initial_delay_seconds"] = int(in.InitialDelaySeconds)
	}

	return []interface{}{att}
}

func flattenHTTPGet(in *k8sv1.HTTPGetAction) []interface{} {
	att := make(map[string]interface{})

	if in.Path != "" {
		att["path"] = in.Path
	}

	switch in.Port.Type {
	case intstr.Int:
		att["port"] = strconv.Itoa(int(in.Port.IntVal))
	case intstr.String:
		att["port"] = in.Port.StrVal
	}

	if in.Scheme != "" {
		att["scheme"] = string(in.Scheme)
	}

	if in.Host != "" {
		att["host"] = in.Host
	}

	if len(in.HTTPHeaders) > 0 {
		att["http_header"] = flattenHTTPHeader(in.HTTPHeaders)
	}

	return []interface{}{att}
}

func flattenHTTPHeader(in []k8sv1.HTTPHeader) []interface{} {
	att := make([]interface{}, 0, len(in))

	for _, h := range in {
		att = append(att, map[string]interface{}{
			"name":  h.Name,
			"value": h.Value,
		})
	}

	return att
}

func flattenTCPSocket(in *k8sv1.TCPSocketAction) []interface{} {
	att := make(map[string]interface{})

	switch in.Port.Type {
	case intstr.Int:
		att["port"] = in.Port.IntValue()

	case intstr.String:
		att["port"] = in.Port.String()

	}

	return []interface{}{att}
}
