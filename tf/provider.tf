terraform {
  required_providers { ovh = { source = "ovh/ovh" } }
}
provider "ovh" { endpoint = "ovh-eu" }
