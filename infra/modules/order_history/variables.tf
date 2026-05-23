variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}
