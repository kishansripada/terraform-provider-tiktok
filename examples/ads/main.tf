terraform {
  required_version = ">= 1.14.0, < 2.0.0"
  required_providers {
    tiktok = {
      source  = "singingbuddy/tiktok"
      version = "0.3.0"
    }
  }
}

provider "tiktok" {
  advertiser_id = "1234567890123456789"
  allow_writes  = false
  allow_delete  = false
}

resource "tiktok_campaign" "traffic" {
  campaign_name    = "Website traffic"
  objective_type   = "TRAFFIC"
  budget_mode      = "BUDGET_MODE_INFINITE"
  operation_status = "DISABLE"
}

resource "tiktok_adgroup" "website" {
  campaign_id         = tiktok_campaign.traffic.id
  adgroup_name        = "Website visitors"
  promotion_type      = "WEBSITE"
  placement_type      = "PLACEMENT_TYPE_NORMAL"
  placements          = ["PLACEMENT_TIKTOK"]
  location_ids        = ["6252001"]
  budget_mode         = "BUDGET_MODE_DAY"
  budget              = 20
  schedule_type       = "SCHEDULE_FROM_NOW"
  schedule_start_time = "2026-10-01 00:00:00"
  optimization_goal   = "CLICK"
  billing_event       = "CPC"
  bid_type            = "BID_TYPE_NO_BID"
  pacing              = "PACING_MODE_SMOOTH"
  operation_status    = "DISABLE"
}

# Replace these references with assets/identity authorized for your advertiser.
# The provider does not upload or publish media.
resource "tiktok_ad" "video" {
  adgroup_id       = tiktok_adgroup.website.id
  ad_name          = "Website video"
  ad_format        = "SINGLE_VIDEO"
  identity_type    = "TT_USER"
  identity_id      = "1234567890123456789"
  video_id         = "your-existing-video-id"
  ad_text          = "Visit our website"
  landing_page_url = "https://example.com"
  operation_status = "DISABLE"
}
