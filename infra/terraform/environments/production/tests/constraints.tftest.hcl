mock_provider "oci" {}

override_data {
  target = data.oci_identity_availability_domains.available
  values = {
    availability_domains = [{ name = "AD-1" }]
  }
}

override_data {
  target = data.oci_core_images.oracle_linux
  values = {
    images = [{ id = "ocid1.image.oc1..test" }]
  }
}

variables {
  region                    = "sa-saopaulo-1"
  tenancy_id                = "ocid1.tenancy.oc1..test"
  compartment_id            = "ocid1.compartment.oc1..test"
  availability_domain_index = 0
  admin_ipv4_cidr           = "203.0.113.10/32"
  cloudflare_ipv4_cidrs     = ["173.245.48.0/20"]
  ssh_public_key_path       = "tests/fixtures/operator.pub"
  mysql_admin_password      = "Test-only-password-9!"
  alarm_email               = "operator@example.com"
}

run "always_free_topology" {
  command = plan

  assert {
    condition     = var.compute_shape == "VM.Standard.E2.1.Micro"
    error_message = "The compute shape escaped the Always Free guard."
  }

  assert {
    condition     = var.mysql_shape == "MySQL.Free"
    error_message = "The MySQL shape escaped the Always Free guard."
  }

  assert {
    condition     = cidrhost(var.database_subnet_cidr, 10) == var.database_private_ip
    error_message = "The database address escaped its private subnet."
  }
}

run "reject_paid_compute_shape" {
  command = plan
  variables {
    compute_shape = "VM.Standard.E4.Flex"
  }
  expect_failures = [var.compute_shape]
}

run "reject_broad_ssh_source" {
  command = plan
  variables {
    admin_ipv4_cidr = "0.0.0.0/0"
  }
  expect_failures = [var.admin_ipv4_cidr]
}
