# tiktok_smart_plus_adgroup

Manage one smart plus adgroup through the TikTok Marketing API.

Import: `terraform import tiktok_smart_plus_adgroup.example "ADVERTISER_ID/RESOURCE_ID"`.

New resources must set `operation_status = "DISABLE"`. Enable in a later apply.
Fields are optional in the Terraform schema to support partial ownership and imports; creation requirements below still apply. Conditional requirements and account eligibility are also enforced by TikTok.

Computed attributes: `id`, `advertiser_id`, `observed_json` (raw last response).

| Attribute | Type | Required to create | Update endpoint accepts |
| --- | --- | --- | --- |
| `adgroup_name` | `string` | Yes | Yes |
| `app_config` | `list(object)` | No | No |
| `app_id` | `string` | No | No |
| `attribution_event_count` | `string` | No | No |
| `bid_price` | `number` | No | Yes |
| `bid_type` | `string` | Yes | No |
| `billing_event` | `string` | Yes | No |
| `budget` | `number` | No | Yes |
| `budget_auto_adjust_strategy` | `string` | No | Yes |
| `budget_mode` | `string` | No | No |
| `campaign_id` | `string` | Yes | No |
| `catalog_authorized_bc_id` | `string` | No | No |
| `catalog_id` | `string` | No | No |
| `click_attribution_window` | `string` | No | No |
| `comment_disabled` | `boolean` | No | Yes |
| `conversion_bid_price` | `number` | No | Yes |
| `custom_conversion_id` | `string` | No | No |
| `dayparting` | `string` | No | Yes |
| `deep_bid_type` | `string` | No | No |
| `deep_funnel_event_source` | `string` | No | No |
| `deep_funnel_event_source_id` | `string` | No | No |
| `deep_funnel_optimization_event` | `string` | No | No |
| `deep_funnel_optimization_status` | `string` | No | No |
| `engaged_view_attribution_window` | `string` | No | No |
| `gaming_ad_compliance_agreement` | `string` | No | No |
| `identity_authorized_bc_id` | `string` | No | No |
| `identity_id` | `string` | No | No |
| `identity_type` | `string` | No | No |
| `is_hfss` | `boolean` | No | Yes |
| `is_lhf_compliance` | `boolean` | No | Yes |
| `message_event_set_id` | `string` | No | No |
| `messaging_app_account_id` | `string` | No | No |
| `messaging_app_type` | `string` | No | No |
| `min_budget` | `number` | No | No |
| `minis_id` | `string` | No | No |
| `movie_premiere_date` | `string` | No | No |
| `native_series_id` | `string` | No | No |
| `operation_status` | `string` | Yes | Yes |
| `optimization_event` | `string` | No | No |
| `optimization_goal` | `string` | Yes | No |
| `phone_info` | `object` | No | No |
| `pixel_id` | `string` | No | No |
| `placement_type` | `string` | No | No |
| `placements` | `list(string)` | No | No |
| `promotion_target_type` | `string` | No | No |
| `promotion_type` | `string` | Yes | No |
| `roas_bid` | `number` | No | Yes |
| `schedule_end_time` | `string` | No | Yes |
| `schedule_start_time` | `string` | Yes | Yes |
| `schedule_type` | `string` | Yes | Yes |
| `share_disabled` | `boolean` | No | Yes |
| `suggestion_audience_enabled` | `boolean` | No | Yes |
| `targeting_optimization_mode` | `string` | No | Yes |
| `targeting_spec` | `object` | Yes | Yes |
| `tiktok_subplacements` | `list(string)` | No | No |
| `vbo_window` | `string` | No | No |
| `video_download_disabled` | `boolean` | No | No |
| `view_attribution_window` | `string` | No | No |
| `zalo_id_type` | `string` | No | No |

Full field descriptions, nested object fields, enum values, and conditional rules are pinned in [objects.json](../../objects.json).

See [lifecycle limitations](../../README.md#lifecycle-and-limitations) before applying changes.
