# tiktok_adgroup

Manage one adgroup through the TikTok Marketing API.

Import: `terraform import tiktok_adgroup.example "ADVERTISER_ID/RESOURCE_ID"`.

New resources must set `operation_status = "DISABLE"`. Enable in a later apply.
Fields are optional in the Terraform schema to support partial ownership and imports; creation requirements below still apply. Conditional requirements and account eligibility are also enforced by TikTok.

Computed attributes: `id`, `advertiser_id`, `observed_json` (raw last response).

| Attribute | Type | Required to create | Update endpoint accepts |
| --- | --- | --- | --- |
| `actions` | `list(object)` | No | Yes |
| `adgroup_app_profile_page_state` | `string` | No | No |
| `adgroup_name` | `string` | Yes | Yes |
| `age_groups` | `list(string)` | No | Yes |
| `app_config` | `list(object)` | No | No |
| `app_id` | `string` | No | No |
| `attribution_event_count` | `string` | No | No |
| `audience_ids` | `list(string)` | No | Yes |
| `audience_rule` | `object` | No | Yes |
| `audience_type` | `string` | No | Yes |
| `automated_keywords_enabled` | `boolean` | No | Yes |
| `bid_display_mode` | `string` | No | No |
| `bid_price` | `number` | No | Yes |
| `bid_type` | `string` | No | Yes |
| `billing_event` | `string` | Yes | No |
| `blocked_pangle_app_ids` | `list(string)` | No | Yes |
| `brand_safety_partner` | `string` | No | No |
| `brand_safety_type` | `string` | No | Yes |
| `budget` | `number` | Yes | Yes |
| `budget_mode` | `string` | Yes | No |
| `campaign_id` | `string` | Yes | No |
| `carrier_ids` | `list(string)` | No | Yes |
| `catalog_authorized_bc_id` | `string` | No | Yes |
| `catalog_id` | `string` | No | No |
| `category_exclusion_ids` | `list(string)` | No | Yes |
| `click_attribution_window` | `string` | No | No |
| `comment_disabled` | `boolean` | No | Yes |
| `contextual_tag_ids` | `list(string)` | No | Yes |
| `conversion_bid_price` | `number` | No | Yes |
| `creative_material_mode` | `string` | No | No |
| `custom_conversion_id` | `string` | No | No |
| `dayparting` | `string` | No | Yes |
| `deep_bid_type` | `string` | No | Yes |
| `deep_cpa_bid` | `number` | No | Yes |
| `deep_funnel_event_source` | `string` | No | Yes |
| `deep_funnel_event_source_id` | `string` | No | Yes |
| `deep_funnel_optimization_event` | `string` | No | Yes |
| `deep_funnel_optimization_status` | `string` | No | Yes |
| `device_model_ids` | `list(string)` | No | Yes |
| `device_price_ranges` | `list(number)` | No | Yes |
| `engaged_view_attribution_window` | `string` | No | No |
| `exclude_age_under_eighteen` | `boolean` | No | Yes |
| `excluded_audience_ids` | `list(string)` | No | Yes |
| `excluded_custom_actions` | `list(object)` | No | Yes |
| `excluded_pangle_audience_package_ids` | `list(string)` | No | Yes |
| `frequency` | `number` | No | Yes |
| `frequency_schedule` | `number` | No | Yes |
| `gender` | `string` | No | Yes |
| `household_income` | `list(string)` | No | Yes |
| `identity_authorized_bc_id` | `string` | No | No |
| `identity_id` | `string` | No | No |
| `identity_type` | `string` | No | No |
| `included_custom_actions` | `list(object)` | No | Yes |
| `included_pangle_audience_package_ids` | `list(string)` | No | Yes |
| `interest_category_ids` | `list(string)` | No | Yes |
| `interest_keyword_ids` | `list(string)` | No | Yes |
| `ios14_targeting` | `string` | No | Yes |
| `is_hfss` | `boolean` | No | Yes |
| `is_lhf_compliance` | `boolean` | No | Yes |
| `isp_ids` | `list(string)` | No | Yes |
| `languages` | `list(string)` | No | Yes |
| `location_ids` | `list(string)` | No | Yes |
| `message_event_set_id` | `string` | No | No |
| `messaging_app_account_id` | `string` | No | No |
| `messaging_app_type` | `string` | No | No |
| `min_android_version` | `string` | No | Yes |
| `min_ios_version` | `string` | No | Yes |
| `network_types` | `list(string)` | No | Yes |
| `next_day_retention` | `number` | No | Yes |
| `operating_systems` | `list(string)` | No | Yes |
| `operation_status` | `string` | Yes | Yes |
| `optimization_event` | `string` | No | No |
| `optimization_goal` | `string` | Yes | No |
| `pacing` | `string` | Yes | Yes |
| `phone_number` | `string` | No | No |
| `phone_region_calling_code` | `string` | No | No |
| `phone_region_code` | `string` | No | No |
| `pixel_id` | `string` | No | No |
| `placement_type` | `string` | No | No |
| `placements` | `list(string)` | No | No |
| `product_source` | `string` | No | No |
| `promotion_target_type` | `string` | No | No |
| `promotion_type` | `string` | No | No |
| `promotion_website_type` | `string` | No | No |
| `purchase_intention_keyword_ids` | `list(string)` | No | Yes |
| `roas_bid` | `number` | No | Yes |
| `saved_audience_id` | `string` | No | Yes |
| `schedule_end_time` | `string` | No | Yes |
| `schedule_start_time` | `string` | Yes | Yes |
| `schedule_type` | `string` | Yes | Yes |
| `search_keywords` | `list(object)` | No | Yes |
| `search_result_enabled` | `boolean` | No | Yes |
| `secondary_optimization_event` | `string` | No | Yes |
| `share_disabled` | `boolean` | No | Yes |
| `shopping_ads_retargeting_actions_days` | `number` | No | Yes |
| `shopping_ads_retargeting_custom_audience_relation` | `string` | No | Yes |
| `shopping_ads_retargeting_type` | `string` | No | Yes |
| `shopping_ads_type` | `string` | No | No |
| `smart_audience_enabled` | `boolean` | No | Yes |
| `smart_interest_behavior_enabled` | `boolean` | No | Yes |
| `spending_power` | `string` | No | Yes |
| `statistic_type` | `string` | No | No |
| `store_authorized_bc_id` | `string` | No | No |
| `store_id` | `string` | No | No |
| `tiktok_subplacements` | `list(string)` | No | No |
| `vbo_window` | `string` | No | No |
| `vertical_sensitivity_id` | `string` | No | Yes |
| `video_download_disabled` | `boolean` | No | No |
| `view_attribution_window` | `string` | No | No |
| `zipcode_ids` | `list(string)` | No | Yes |

Full field descriptions, nested object fields, enum values, and conditional rules are pinned in [objects.json](../../objects.json).

See [lifecycle limitations](../../README.md#lifecycle-and-limitations) before applying changes.
