# tiktok_query

Read one response from an allowlisted TikTok GET endpoint. Supported operations
are listed in [coverage](../coverage.md); exact parameter schemas are pinned in
[queries.json](../../queries.json). No arbitrary URL or write operation is accepted.

```hcl
data "tiktok_query" "account" {
  operation = "advertiser_info_get"
}

data "tiktok_query" "videos" {
  operation = "file_video_ad_info_get"
  parameters_json = jsonencode({ video_ids = ["your-existing-video-id"] })
}

data "tiktok_query" "report" {
  operation = "report_integrated_get"
  parameters_json = jsonencode({
    report_type = "BASIC"
    data_level  = "AUCTION_CAMPAIGN"
    dimensions  = ["campaign_id"]
    metrics     = ["spend", "impressions"]
    start_date  = "2026-09-01"
    end_date    = "2026-09-02"
    page        = 1
    page_size   = 100
  })
}
```

- `operation` (required string): exact allowlisted operation name.
- `parameters_json` (optional string): request JSON object; defaults to `{}`.
  Known parameters, required fields, types, and enums are validated. Where supported,
  the configured advertiser is injected. Advertiser overrides must match it.
  Business Center queries require the appropriate `bc_id` and token permissions.
- `result_json` (computed, sensitive string): the response's `data` object.
  Use `jsondecode()` to access fields. Sensitive values are still stored in state.

Each read makes **one request**. Pagination is explicit and response metadata is
preserved. A page is not advertised as the complete collection. For all campaigns,
ad groups, and ads with checked pagination, use `tiktok_inventory` instead.
Refresh/plan may rerun these GET requests. Reports are observations, not desired
configuration; this data source does not provide a reporting history store.
