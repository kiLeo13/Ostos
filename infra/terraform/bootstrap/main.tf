data "oci_objectstorage_namespace" "current" {
  compartment_id = var.compartment_id
}

resource "oci_objectstorage_bucket" "terraform_state" {
  compartment_id = var.compartment_id
  namespace      = data.oci_objectstorage_namespace.current.namespace
  name           = var.state_bucket_name
  access_type    = "NoPublicAccess"
  storage_tier   = "Standard"
  versioning     = "Enabled"

  freeform_tags = {
    application = "ostos"
    purpose     = "terraform-state"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "oci_identity_policy" "terraform_state" {
  compartment_id = var.policy_compartment_id
  name           = "ostos-terraform-state"
  description    = "Restrict the Ostos Terraform operators to its versioned state bucket."
  statements = [
    "Allow group id ${var.terraform_state_group_id} to read buckets in compartment id ${var.compartment_id} where target.bucket.name = '${var.state_bucket_name}'",
    "Allow group id ${var.terraform_state_group_id} to manage objects in compartment id ${var.compartment_id} where target.bucket.name = '${var.state_bucket_name}'",
  ]
}
