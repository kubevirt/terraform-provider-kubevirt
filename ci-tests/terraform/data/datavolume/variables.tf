variable "dv-from-http-name" {
  type = string
}

variable "dv-from-pvc-name" {
  type = string
}

variable "namespace" {
  type = string
}

variable "url" {
  type = string
}

variable "labels" {
  type = map(string)

  description = <<EOF
(optional) Labels to be applied to created resources.

Example: `{ "key" = "value", "foo" = "bar" }`
EOF

  default = {}
}
