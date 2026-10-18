data "oci_identity_availability_domains" "available" {
  compartment_id = var.tenancy_id
}

data "oci_core_images" "oracle_linux" {
  compartment_id           = var.compartment_id
  operating_system         = "Oracle Linux"
  operating_system_version = "9"
  shape                    = var.compute_shape
  sort_by                  = "TIMECREATED"
  sort_order               = "DESC"
}

locals {
  availability_domain = try(data.oci_identity_availability_domains.available.availability_domains[var.availability_domain_index].name, "")
  image_id            = var.compute_image_id != null ? var.compute_image_id : try(data.oci_core_images.oracle_linux.images[0].id, "")
}

check "selected_availability_domain_exists" {
  assert {
    condition     = local.availability_domain != ""
    error_message = "availability_domain_index does not identify an AD returned by OCI."
  }
}

check "oracle_linux_image_exists" {
  assert {
    condition     = local.image_id != ""
    error_message = "No Oracle Linux 9 image compatible with the selected compute shape was found; pin compute_image_id."
  }
}

module "network" {
  source = "../../modules/network"

  compartment_id        = var.compartment_id
  name_prefix           = "ostos-production"
  vcn_cidr              = var.vcn_cidr
  public_subnet_cidr    = var.public_subnet_cidr
  database_subnet_cidr  = var.database_subnet_cidr
  admin_ipv4_cidr       = var.admin_ipv4_cidr
  cloudflare_ipv4_cidrs = var.cloudflare_ipv4_cidrs
}

module "compute" {
  source = "../../modules/compute"

  availability_domain = local.availability_domain
  compartment_id      = var.compartment_id
  display_name        = "ostos-api-production"
  image_id            = local.image_id
  shape               = var.compute_shape
  subnet_id           = module.network.public_subnet_id
  nsg_ids             = [module.network.compute_nsg_id, module.network.mysql_client_nsg_id]
  private_ip          = var.compute_private_ip
  ssh_public_key      = file(var.ssh_public_key_path)
  snowflake_node_id   = var.snowflake_node_id
}

module "database" {
  source = "../../modules/database"

  availability_domain = local.availability_domain
  compartment_id      = var.compartment_id
  display_name        = "ostos-mysql-production"
  shape_name          = var.mysql_shape
  subnet_id           = module.network.database_subnet_id
  nsg_ids             = [module.network.mysql_nsg_id]
  private_ip          = var.database_private_ip
  admin_username      = var.mysql_admin_username
  admin_password      = var.mysql_admin_password
}

resource "oci_ons_notification_topic" "operations" {
  compartment_id = var.compartment_id
  name           = "ostos-production-operations"
  description    = "Low-volume production infrastructure alarms for Ostos."
}

resource "oci_ons_subscription" "operations_email" {
  compartment_id = var.compartment_id
  endpoint       = var.alarm_email
  protocol       = "EMAIL"
  topic_id       = oci_ons_notification_topic.operations.id
}

resource "oci_monitoring_alarm" "instance_unreachable" {
  compartment_id               = var.compartment_id
  destinations                 = [oci_ons_notification_topic.operations.id]
  display_name                 = "ostos-api-instance-unreachable"
  is_enabled                   = true
  metric_compartment_id        = var.compartment_id
  namespace                    = "oci_compute_infrastructure_health"
  pending_duration             = "PT5M"
  query                        = "instance_status[1m]{resourceId = \"${module.compute.instance_id}\"}.max() > 0"
  resolution                   = "1m"
  severity                     = "CRITICAL"
  body                         = "The Ostos production VM has been unavailable for at least five minutes."
  message_format               = "ONS_OPTIMIZED"
  repeat_notification_duration = "PT2H"
}
