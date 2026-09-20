# tiktok_campaign

Manage one campaign through the TikTok Marketing API.

Import: `terraform import tiktok_campaign.example "ADVERTISER_ID/RESOURCE_ID"`.

New resources must set `operation_status = "DISABLE"`. Enable in a later apply.
Fields are optional in the Terraform schema to support partial ownership and imports; creation requirements below still apply. Conditional requirements and account eligibility are also enforced by TikTok.

Computed attributes: `id`, `advertiser_id`.

| Attribute | Type | Required to create | Update endpoint accepts |
| --- | --- | --- | --- |
| `app_id` | `string` | No | No |
| `app_promotion_type` | `string` | No | No |
| `budget` | `number` | No | Yes |
| `budget_mode` | `string` | No | No |
| `budget_optimize_on` | `boolean` | No | No |
| `campaign_app_profile_page_state` | `string` | No | No |
| `campaign_name` | `string` | Yes | Yes |
| `campaign_product_source` | `string` | No | No |
| `campaign_type` | `string` | No | No |
| `catalog_enabled` | `boolean` | No | No |
| `disable_skan_campaign` | `boolean` | No | No |
| `is_advanced_dedicated_campaign` | `boolean` | No | No |
| `is_search_campaign` | `boolean` | No | No |
| `objective_type` | `string` | Yes | No |
| `operation_status` | `string` | Yes | Yes |
| `po_number` | `string` | No | Yes |
| `postback_window_mode` | `string` | No | No |
| `rf_campaign_type` | `string` | No | No |
| `rta_bid_enabled` | `boolean` | No | No |
| `rta_id` | `string` | No | No |
| `rta_product_selection_enabled` | `boolean` | No | No |
| `sales_destination` | `string` | No | No |
| `special_industries` | `list(string)` | No | Yes |
| `virtual_objective_type` | `string` | No | No |

Full field descriptions, nested object fields, enum values, and conditional rules are pinned in [campaigns.json](../../campaigns.json).

See [lifecycle limitations](../../README.md#lifecycle-and-limitations) before applying changes.
