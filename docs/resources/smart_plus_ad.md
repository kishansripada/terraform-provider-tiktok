# tiktok_smart_plus_ad

Manage one smart plus ad through the TikTok Marketing API.

Import: `terraform import tiktok_smart_plus_ad.example "ADVERTISER_ID/RESOURCE_ID"`.

New resources must set `operation_status = "DISABLE"`. Enable in a later apply.
Fields are optional in the Terraform schema to support partial ownership and imports; creation requirements below still apply. Conditional requirements and account eligibility are also enforced by TikTok.

Computed attributes: `id`, `advertiser_id`, `observed_json` (raw last response).

| Attribute | Type | Required to create | Update endpoint accepts |
| --- | --- | --- | --- |
| `ad_configuration` | `object` | No | Yes |
| `ad_name` | `string` | No | Yes |
| `ad_text_list` | `list(object)` | No | Yes |
| `adgroup_id` | `string` | Yes | No |
| `auto_message_list` | `list(object)` | No | No |
| `call_to_action_list` | `list(object)` | No | Yes |
| `creative_list` | `list(object)` | No | Yes |
| `custom_product_page_list` | `list(object)` | No | Yes |
| `deeplink_list` | `list(object)` | No | Yes |
| `disclaimer` | `object` | No | Yes |
| `interactive_add_on_list` | `list(object)` | No | Yes |
| `landing_page_url_list` | `list(object)` | No | Yes |
| `operation_status` | `string` | Yes | Yes |
| `page_list` | `list(object)` | No | Yes |
| `playable_list` | `list(object)` | No | Yes |

Full field descriptions, nested object fields, enum values, and conditional rules are pinned in [objects.json](../../objects.json).

See [lifecycle limitations](../../README.md#lifecycle-and-limitations) before applying changes.
