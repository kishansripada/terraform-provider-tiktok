# tiktok_ad

Manage one ad through the TikTok Marketing API.

Import: `terraform import tiktok_ad.example "ADVERTISER_ID/RESOURCE_ID"`.

New resources must set `operation_status = "DISABLE"`. Enable in a later apply.
Fields are optional in the Terraform schema to support partial ownership and imports; creation requirements below still apply. Conditional requirements and account eligibility are also enforced by TikTok.

Computed attributes: `id`, `advertiser_id`, `observed_json` (raw last response).

| Attribute | Type | Required to create | Update endpoint accepts |
| --- | --- | --- | --- |
| `ad_format` | `string` | Yes | Yes |
| `ad_name` | `string` | Yes | Yes |
| `ad_text` | `string` | No | Yes |
| `ad_texts` | `list(string)` | No | Yes |
| `adgroup_id` | `string` | Yes | No |
| `aigc_disclosure_type` | `string` | No | Yes |
| `app_name` | `string` | No | Yes |
| `auto_disclaimer_types` | `list(string)` | No | No |
| `auto_message_id` | `string` | No | No |
| `avatar_icon_web_uri` | `string` | No | Yes |
| `brand_safety_postbid_partner` | `string` | No | Yes |
| `brand_safety_vast_url` | `string` | No | Yes |
| `call_to_action` | `string` | No | Yes |
| `call_to_action_id` | `string` | No | Yes |
| `card_id` | `string` | No | Yes |
| `carousel_image_index` | `integer` | No | Yes |
| `catalog_id` | `string` | No | No |
| `click_tracking_url` | `string` | No | Yes |
| `cpp_url` | `string` | No | Yes |
| `creative_authorized` | `boolean` | No | Yes |
| `creative_auto_enhancement_strategy_list` | `list(string)` | No | No |
| `creative_type` | `string` | No | Yes |
| `dark_post_status` | `string` | No | Yes |
| `deeplink` | `string` | No | Yes |
| `deeplink_format_type` | `string` | No | No |
| `deeplink_type` | `string` | No | Yes |
| `deeplink_utm_params` | `list(object)` | No | Yes |
| `disclaimer_clickable_texts` | `list(object)` | No | Yes |
| `disclaimer_text` | `object` | No | Yes |
| `disclaimer_type` | `string` | No | Yes |
| `display_name` | `string` | No | Yes |
| `dynamic_destination` | `string` | No | Yes |
| `dynamic_format` | `string` | No | Yes |
| `end_card_cta` | `string` | No | Yes |
| `fallback_type` | `string` | No | Yes |
| `identity_authorized_bc_id` | `string` | No | Yes |
| `identity_id` | `string` | Yes | Yes |
| `identity_type` | `string` | Yes | Yes |
| `image_ids` | `list(string)` | No | Yes |
| `impression_tracking_url` | `string` | No | Yes |
| `instant_product_page_used` | `boolean` | No | Yes |
| `item_duet_status` | `string` | No | Yes |
| `item_group_ids` | `list(string)` | No | Yes |
| `item_stitch_status` | `string` | No | Yes |
| `landing_page_url` | `string` | No | Yes |
| `music_id` | `string` | No | Yes |
| `operation_status` | `string` | Yes | Yes |
| `page_id` | `number` | No | Yes |
| `page_image_index` | `integer` | No | Yes |
| `phone_number` | `string` | No | No |
| `phone_region_calling_code` | `string` | No | No |
| `phone_region_code` | `string` | No | No |
| `playable_url` | `string` | No | Yes |
| `product_display_field_list` | `list(string)` | No | No |
| `product_set_id` | `string` | No | Yes |
| `product_specific_type` | `string` | No | Yes |
| `promotional_music_disabled` | `boolean` | No | Yes |
| `schedule_id` | `string` | No | No |
| `shopping_ads_deeplink_type` | `string` | No | Yes |
| `shopping_ads_fallback_type` | `string` | No | Yes |
| `shopping_ads_video_package_id` | `string` | No | Yes |
| `showcase_products` | `list(object)` | No | No |
| `sku_ids` | `list(string)` | No | Yes |
| `tiktok_item_id` | `string` | No | Yes |
| `tiktok_page_category` | `string` | No | No |
| `tracking_app_id` | `string` | No | Yes |
| `tracking_message_event_set_id` | `string` | No | No |
| `tracking_offline_event_set_ids` | `list(string)` | No | Yes |
| `tracking_pixel_id` | `number` | No | Yes |
| `utm_params` | `list(object)` | No | Yes |
| `vast_moat_enabled` | `boolean` | No | Yes |
| `vehicle_ids` | `list(string)` | No | No |
| `vertical_video_strategy` | `string` | No | Yes |
| `video_id` | `string` | No | Yes |
| `video_view_tracking_url` | `string` | No | Yes |
| `viewability_postbid_partner` | `string` | No | Yes |
| `viewability_vast_url` | `string` | No | Yes |

Full field descriptions, nested object fields, enum values, and conditional rules are pinned in [objects.json](../../objects.json).

See [lifecycle limitations](../../README.md#lifecycle-and-limitations) before applying changes.
