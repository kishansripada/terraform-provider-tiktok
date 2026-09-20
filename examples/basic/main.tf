terraform {
  required_version = ">= 1.14.0, < 2.0.0"
  required_providers {
    tiktok = {
      source  = "singingbuddy/tiktok"
      version = "0.3.0"
    }
  }
}

variable "advertiser_id" {
  type        = string
  description = "Your TikTok advertiser ID."
}

variable "allow_writes" {
  type        = bool
  default     = false
  description = "Enable only after reviewing the plan."
}

provider "tiktok" {
  advertiser_id = var.advertiser_id
  allow_writes  = var.allow_writes
}

resource "tiktok_campaign" "example" {
  campaign_name    = "Terraform example"
  objective_type   = "TRAFFIC"
  budget_mode      = "BUDGET_MODE_DAY"
  budget           = 50
  operation_status = "DISABLE"

  lifecycle {
    prevent_destroy = true
  }
}

data "tiktok_inventory" "account" {}
