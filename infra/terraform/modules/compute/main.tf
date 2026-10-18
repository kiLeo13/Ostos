resource "oci_core_instance" "this" {
  availability_domain = var.availability_domain
  compartment_id      = var.compartment_id
  display_name        = var.display_name
  shape               = var.shape

  agent_config {
    is_management_disabled = false
    is_monitoring_disabled = false
    plugins_config {
      desired_state = "ENABLED"
      name          = "Compute Instance Monitoring"
    }
  }

  create_vnic_details {
    assign_public_ip = false
    display_name     = "${var.display_name}-primary"
    hostname_label   = "api"
    nsg_ids          = var.nsg_ids
    private_ip       = var.private_ip
    subnet_id        = var.subnet_id
  }

  source_details {
    source_id   = var.image_id
    source_type = "image"
  }

  metadata = {
    ssh_authorized_keys = trimspace(var.ssh_public_key)
  }

  freeform_tags = {
    application       = "ostos"
    snowflake_node_id = tostring(var.snowflake_node_id)
  }

  lifecycle {
    precondition {
      condition     = var.shape == "VM.Standard.E2.1.Micro"
      error_message = "Ostos production permits only the Always Free VM.Standard.E2.1.Micro shape."
    }
  }
}

data "oci_core_vnic_attachments" "this" {
  compartment_id = var.compartment_id
  instance_id    = oci_core_instance.this.id
}

data "oci_core_private_ips" "primary" {
  vnic_id = data.oci_core_vnic_attachments.this.vnic_attachments[0].vnic_id
}

resource "oci_core_public_ip" "this" {
  compartment_id = var.compartment_id
  display_name   = "${var.display_name}-reserved"
  lifetime       = "RESERVED"
  private_ip_id  = data.oci_core_private_ips.primary.private_ips[0].id

  freeform_tags = {
    application = "ostos"
  }

  lifecycle {
    prevent_destroy = true
  }
}
