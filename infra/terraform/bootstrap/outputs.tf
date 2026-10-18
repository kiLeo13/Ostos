output "backend_bucket" {
  value       = oci_objectstorage_bucket.terraform_state.name
  description = "Bucket value for backend.production.tfbackend."
}

output "backend_namespace" {
  value       = data.oci_objectstorage_namespace.current.namespace
  description = "Namespace value for backend.production.tfbackend."
}
