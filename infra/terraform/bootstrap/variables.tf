variable "region" {
  description = "OCI home region containing the state bucket."
  type        = string
}

variable "oci_profile" {
  description = "Non-secret profile name from the operator's OCI configuration."
  type        = string
  default     = "DEFAULT"
}

variable "compartment_id" {
  description = "Compartment OCID for the state bucket and its least-privilege policy."
  type        = string
}

variable "state_bucket_name" {
  description = "Globally unique name inside the tenancy Object Storage namespace."
  type        = string
}

variable "terraform_state_group_id" {
  description = "OCID of the existing OCI IAM group allowed to access only this state bucket's objects."
  type        = string
}

variable "policy_compartment_id" {
  description = "Compartment in which to create the IAM policy; normally the tenancy OCID."
  type        = string
}
