# API coverage

This release has **6 managed resource types** and **104 allowlisted read-only query operations**, plus the fully paginated six-collection inventory. It does **not** cover the entire TikTok API.

The table audits the 379-operation MCP snapshot retrieved on 2026-09-20. That snapshot is not a complete OpenAPI response specification or a guarantee of all TikTok endpoints. Routes for generic queries are checked against the pinned official SDK in [spec/routes.json](../spec/routes.json).

“Resource” means the operation participates in a managed lifecycle. “Query” means one GET request through `tiktok_query`, with explicit parameters and pagination. “Not implemented” means no provider execution support; having a schema in the snapshot does not implement it.

Unimplemented areas include persistent catalog/audience/pixel/Business Center writes as well as one-off payments, messages, event delivery, and media uploads. These need separate lifecycle designs and response-contract verification; they are not hidden behind a generic write escape hatch.

| Operation | Support | Query route |
| --- | --- | --- |
| `ad_audience_size_estimate` | Not implemented | — |
| `ad_create` | Resource | — |
| `ad_get` | Resource, Query | `/open_api/v1.3/ad/get/` |
| `ad_review_info_get` | Not implemented | — |
| `ad_status_update` | Resource | — |
| `ad_update` | Resource | — |
| `adgroup_appeal` | Not implemented | — |
| `adgroup_budget_update` | Not implemented | — |
| `adgroup_create` | Resource | — |
| `adgroup_get` | Resource, Query | `/open_api/v1.3/adgroup/get/` |
| `adgroup_quota_get` | Query | `/open_api/v1.3/adgroup/quota/` |
| `adgroup_review_info_get` | Not implemented | — |
| `adgroup_rf_create` | Not implemented | — |
| `adgroup_rf_estimated_info_get` | Not implemented | — |
| `adgroup_rf_update` | Not implemented | — |
| `adgroup_status_update` | Resource | — |
| `adgroup_update` | Resource | — |
| `advertiser_balance_get` | Query | `/open_api/v1.3/advertiser/balance/get/` |
| `advertiser_info_get` | Query | `/open_api/v1.3/advertiser/info/` |
| `advertiser_transaction_get` | Query | `/open_api/v1.3/advertiser/transaction/get/` |
| `advertiser_update` | Not implemented | — |
| `app_create` | Not implemented | — |
| `app_info_get` | Query | `/open_api/v1.3/app/info/` |
| `app_list_get` | Query | `/open_api/v1.3/app/list/` |
| `app_optimization_event_get` | Query | `/open_api/v1.3/app/optimization_event/` |
| `app_optimization_event_retargeting_get` | Query | `/open_api/v1.3/app/optimization_event/retargeting/` |
| `app_update` | Not implemented | — |
| `asset_bind_quota_get` | Not implemented | — |
| `audience_insight_info_get` | Not implemented | — |
| `audience_insight_overlap_get` | Query | `/open_api/v1.3/audience/insight/overlap/` |
| `audience_segment_operate` | Not implemented | — |
| `auth_advertiser_get` | Not implemented | — |
| `bc_account_budget_changelog_get` | Not implemented | — |
| `bc_account_cost_get` | Not implemented | — |
| `bc_account_transaction_get` | Query | `/open_api/v1.3/bc/account/transaction/get/` |
| `bc_advertiser_attribute_get` | Not implemented | — |
| `bc_advertiser_create` | Not implemented | — |
| `bc_advertiser_disable` | Not implemented | — |
| `bc_advertiser_qualification_get` | Not implemented | — |
| `bc_advertiser_unionpay_info_check` | Not implemented | — |
| `bc_asset_account_authorization_get` | Not implemented | — |
| `bc_asset_admin_delete` | Not implemented | — |
| `bc_asset_admin_get` | Query | `/open_api/v1.3/bc/asset/admin/get/` |
| `bc_asset_advertiser_assign` | Not implemented | — |
| `bc_asset_advertiser_assigned_get` | Not implemented | — |
| `bc_asset_advertiser_unassign` | Not implemented | — |
| `bc_asset_assign` | Not implemented | — |
| `bc_asset_get` | Query | `/open_api/v1.3/bc/asset/get/` |
| `bc_asset_group_create` | Not implemented | — |
| `bc_asset_group_delete` | Not implemented | — |
| `bc_asset_group_get` | Query | `/open_api/v1.3/bc/asset_group/get/` |
| `bc_asset_group_list` | Query | `/open_api/v1.3/bc/asset_group/list/` |
| `bc_asset_group_update` | Not implemented | — |
| `bc_asset_member_get` | Query | `/open_api/v1.3/bc/asset/member/get/` |
| `bc_asset_partner_get` | Query | `/open_api/v1.3/bc/asset/partner/get/` |
| `bc_asset_unassign` | Not implemented | — |
| `bc_balance_get` | Query | `/open_api/v1.3/bc/balance/get/` |
| `bc_billing_group_advertiser_list` | Not implemented | — |
| `bc_billing_group_create` | Not implemented | — |
| `bc_billing_group_get` | Query | `/open_api/v1.3/bc/billing_group/get/` |
| `bc_billing_group_update` | Not implemented | — |
| `bc_get` | Query | `/open_api/v1.3/bc/get/` |
| `bc_invoice_get` | Not implemented | — |
| `bc_invoice_task_create` | Not implemented | — |
| `bc_invoice_task_get` | Not implemented | — |
| `bc_invoice_task_list` | Not implemented | — |
| `bc_invoice_unpaid_get` | Query | `/open_api/v1.3/bc/invoice/unpaid/get/` |
| `bc_member_delete` | Not implemented | — |
| `bc_member_get` | Query | `/open_api/v1.3/bc/member/get/` |
| `bc_member_invite` | Not implemented | — |
| `bc_member_update` | Not implemented | — |
| `bc_partner_add` | Not implemented | — |
| `bc_partner_asset_delete` | Not implemented | — |
| `bc_partner_asset_get` | Query | `/open_api/v1.3/bc/partner/asset/get/` |
| `bc_partner_delete` | Not implemented | — |
| `bc_partner_get` | Query | `/open_api/v1.3/bc/partner/get/` |
| `bc_pixel_link_get` | Query | `/open_api/v1.3/bc/pixel/link/get/` |
| `bc_pixel_link_update` | Not implemented | — |
| `bc_pixel_transfer` | Not implemented | — |
| `bc_transaction_get` | Query | `/open_api/v1.3/bc/transaction/get/` |
| `bc_transfer` | Not implemented | — |
| `blockedword_check` | Query | `/open_api/v1.3/blockedword/check/` |
| `blockedword_create` | Not implemented | — |
| `blockedword_delete` | Not implemented | — |
| `blockedword_list_get` | Query | `/open_api/v1.3/blockedword/list/` |
| `blockedword_task_check` | Query | `/open_api/v1.3/blockedword/task/check/` |
| `blockedword_task_create` | Not implemented | — |
| `blockedword_task_download` | Not implemented | — |
| `blockedword_update` | Not implemented | — |
| `campaign_copy_task_check` | Not implemented | — |
| `campaign_copy_task_create` | Not implemented | — |
| `campaign_create` | Resource | — |
| `campaign_get` | Resource, Query | `/open_api/v1.3/campaign/get/` |
| `campaign_gmv_max_create` | Not implemented | — |
| `campaign_gmv_max_info_get` | Query | `/open_api/v1.3/campaign/gmv_max/info/` |
| `campaign_gmv_max_session_create` | Not implemented | — |
| `campaign_gmv_max_session_delete` | Not implemented | — |
| `campaign_gmv_max_session_get` | Query | `/open_api/v1.3/campaign/gmv_max/session/get/` |
| `campaign_gmv_max_session_list_get` | Query | `/open_api/v1.3/campaign/gmv_max/session/list/` |
| `campaign_gmv_max_session_update` | Not implemented | — |
| `campaign_gmv_max_update` | Not implemented | — |
| `campaign_label_get` | Not implemented | — |
| `campaign_quota_info_get` | Not implemented | — |
| `campaign_status_update` | Resource | — |
| `campaign_update` | Resource | — |
| `catalog_available_country_get` | Query | `/open_api/v1.3/catalog/available_country/get/` |
| `catalog_capitalize_migrate` | Not implemented | — |
| `catalog_create` | Not implemented | — |
| `catalog_delete` | Not implemented | — |
| `catalog_eventsource_bind` | Not implemented | — |
| `catalog_eventsource_bind_get` | Query | `/open_api/v1.3/catalog/eventsource_bind/get/` |
| `catalog_eventsource_unbind` | Not implemented | — |
| `catalog_feed_create` | Not implemented | — |
| `catalog_feed_delete` | Not implemented | — |
| `catalog_feed_get` | Query | `/open_api/v1.3/catalog/feed/get/` |
| `catalog_feed_log_get` | Query | `/open_api/v1.3/catalog/feed/log/` |
| `catalog_feed_switch_update` | Not implemented | — |
| `catalog_feed_update` | Not implemented | — |
| `catalog_get` | Query | `/open_api/v1.3/catalog/get/` |
| `catalog_insight_category_get` | Not implemented | — |
| `catalog_insight_filter_get` | Not implemented | — |
| `catalog_insight_product_get` | Not implemented | — |
| `catalog_lexicon_get` | Query | `/open_api/v1.3/catalog/lexicon/get/` |
| `catalog_location_currency_get` | Query | `/open_api/v1.3/catalog/location_currency/get/` |
| `catalog_overview_get` | Query | `/open_api/v1.3/catalog/overview/` |
| `catalog_product_delete` | Not implemented | — |
| `catalog_product_file_upload` | Not implemented | — |
| `catalog_product_get` | Not implemented | — |
| `catalog_product_log_get` | Query | `/open_api/v1.3/catalog/product/log/` |
| `catalog_product_update` | Not implemented | — |
| `catalog_product_upload` | Not implemented | — |
| `catalog_set_create` | Not implemented | — |
| `catalog_set_delete` | Not implemented | — |
| `catalog_set_get` | Query | `/open_api/v1.3/catalog/set/get/` |
| `catalog_set_product_get` | Query | `/open_api/v1.3/catalog/set/product/get/` |
| `catalog_set_update` | Not implemented | — |
| `catalog_update` | Not implemented | — |
| `catalog_video_delete` | Not implemented | — |
| `catalog_video_file_upload` | Not implemented | — |
| `catalog_video_get` | Query | `/open_api/v1.3/catalog/video/get/` |
| `catalog_video_log_get` | Not implemented | — |
| `catalog_video_package_get` | Not implemented | — |
| `changelog_get` | Not implemented | — |
| `changelog_task_check` | Not implemented | — |
| `changelog_task_create` | Not implemented | — |
| `changelog_task_download` | Not implemented | — |
| `comment_delete` | Not implemented | — |
| `comment_list_get` | Query | `/open_api/v1.3/comment/list/` |
| `comment_post_create` | Not implemented | — |
| `comment_reference_get` | Query | `/open_api/v1.3/comment/reference/` |
| `comment_status_update` | Not implemented | — |
| `comment_task_check` | Query | `/open_api/v1.3/comment/task/check/` |
| `comment_task_create` | Not implemented | — |
| `comment_task_download` | Not implemented | — |
| `creative_ads_preview_create` | Not implemented | — |
| `creative_asset_delete` | Not implemented | — |
| `creative_asset_share_get` | Not implemented | — |
| `creative_auto_message_create` | Not implemented | — |
| `creative_auto_message_get` | Not implemented | — |
| `creative_cta_recommend_get` | Not implemented | — |
| `creative_fatigue_get` | Not implemented | — |
| `creative_image_edit_get` | Not implemented | — |
| `creative_portfolio_create` | Not implemented | — |
| `creative_portfolio_delete` | Not implemented | — |
| `creative_portfolio_get` | Query | `/open_api/v1.3/creative/portfolio/get/` |
| `creative_portfolio_list_get` | Query | `/open_api/v1.3/creative/portfolio/list/` |
| `creative_report_get` | Not implemented | — |
| `creative_smart_text_get` | Not implemented | — |
| `crm_create` | Not implemented | — |
| `crm_list_get` | Not implemented | — |
| `ctm_message_event_set_get` | Not implemented | — |
| `custom_conversion_create` | Not implemented | — |
| `custom_conversion_delete` | Not implemented | — |
| `custom_conversion_get` | Not implemented | — |
| `custom_conversion_list_get` | Not implemented | — |
| `custom_conversion_update` | Not implemented | — |
| `diagnostic_catalog_eventsource_issue_get` | Not implemented | — |
| `diagnostic_catalog_eventsource_metric_get` | Not implemented | — |
| `diagnostic_catalog_get` | Not implemented | — |
| `diagnostic_catalog_product_task_create` | Not implemented | — |
| `diagnostic_catalog_product_task_get` | Not implemented | — |
| `dmp_custom_audience_apply` | Not implemented | — |
| `dmp_custom_audience_apply_log_get` | Query | `/open_api/v1.3/dmp/custom_audience/apply/log/` |
| `dmp_custom_audience_create` | Not implemented | — |
| `dmp_custom_audience_delete` | Not implemented | — |
| `dmp_custom_audience_get` | Query | `/open_api/v1.3/dmp/custom_audience/get/` |
| `dmp_custom_audience_list_get` | Query | `/open_api/v1.3/dmp/custom_audience/list/` |
| `dmp_custom_audience_lookalike_create` | Not implemented | — |
| `dmp_custom_audience_lookalike_update` | Not implemented | — |
| `dmp_custom_audience_rule_create` | Not implemented | — |
| `dmp_custom_audience_share_cancel` | Not implemented | — |
| `dmp_custom_audience_share_get` | Not implemented | — |
| `dmp_custom_audience_share_log_get` | Query | `/open_api/v1.3/dmp/custom_audience/share/log/` |
| `dmp_custom_audience_update` | Not implemented | — |
| `dmp_saved_audience_create` | Not implemented | — |
| `dmp_saved_audience_delete` | Not implemented | — |
| `dmp_saved_audience_list_get` | Query | `/open_api/v1.3/dmp/saved_audience/list/` |
| `file_image_ad_info_get` | Query | `/open_api/v1.3/file/image/ad/info/` |
| `file_image_ad_search` | Not implemented | — |
| `file_image_ad_update` | Not implemented | — |
| `file_image_ad_upload` | Not implemented | — |
| `file_music_get` | Not implemented | — |
| `file_music_upload` | Not implemented | — |
| `file_name_check` | Not implemented | — |
| `file_temporarily_upload` | Not implemented | — |
| `file_video_ad_info_get` | Query | `/open_api/v1.3/file/video/ad/info/` |
| `file_video_ad_search` | Query | `/open_api/v1.3/file/video/ad/search/` |
| `file_video_ad_update` | Not implemented | — |
| `file_video_ad_upload` | Not implemented | — |
| `file_video_suggestcover_get` | Not implemented | — |
| `gmv_max_bid_recommend_get` | Query | `/open_api/v1.3/gmv_max/bid/recommend/` |
| `gmv_max_campaign_get` | Not implemented | — |
| `gmv_max_creative_update` | Not implemented | — |
| `gmv_max_exclusive_authorization_create` | Not implemented | — |
| `gmv_max_exclusive_authorization_get` | Query | `/open_api/v1.3/gmv_max/exclusive_authorization/get/` |
| `gmv_max_identity_get` | Query | `/open_api/v1.3/gmv_max/identity/get/` |
| `gmv_max_occupied_custom_shop_ads_list_check` | Not implemented | — |
| `gmv_max_report_get` | Query | `/open_api/v1.3/gmv_max/report/get/` |
| `gmv_max_store_list_get` | Query | `/open_api/v1.3/gmv_max/store/list/` |
| `gmv_max_store_shop_ad_usage_check_check` | Not implemented | — |
| `gmv_max_video_get` | Query | `/open_api/v1.3/gmv_max/video/get/` |
| `identity_get` | Query | `/open_api/v1.3/identity/get/` |
| `identity_info_get` | Not implemented | — |
| `identity_live_get` | Not implemented | — |
| `identity_music_authorization_get` | Not implemented | — |
| `identity_native_series_get` | Not implemented | — |
| `identity_video_get` | Not implemented | — |
| `identity_video_info_get` | Query | `/open_api/v1.3/identity/video/info/` |
| `imageurl_to_base64` | Not implemented | — |
| `lead_field_get` | Not implemented | — |
| `lead_get` | Not implemented | — |
| `minis_get` | Not implemented | — |
| `offline_create` | Not implemented | — |
| `offline_delete` | Not implemented | — |
| `offline_get` | Query | `/open_api/v1.3/offline/get/` |
| `offline_update` | Not implemented | — |
| `optimizer_rule_batch_bind_get` | Not implemented | — |
| `optimizer_rule_create` | Not implemented | — |
| `optimizer_rule_get` | Query | `/open_api/v1.3/optimizer/rule/get/` |
| `optimizer_rule_list_get` | Query | `/open_api/v1.3/optimizer/rule/list/` |
| `optimizer_rule_result_get` | Query | `/open_api/v1.3/optimizer/rule/result/get/` |
| `optimizer_rule_result_list_get` | Query | `/open_api/v1.3/optimizer/rule/result/list/` |
| `optimizer_rule_update` | Not implemented | — |
| `optimizer_rule_update_status_update` | Not implemented | — |
| `page_field_get` | Not implemented | — |
| `page_get` | Not implemented | — |
| `page_lead_task_create` | Not implemented | — |
| `page_lead_task_download` | Not implemented | — |
| `page_library_get` | Not implemented | — |
| `page_library_transfer` | Not implemented | — |
| `pangle_audience_package_get` | Query | `/open_api/v1.3/pangle_audience_package/get/` |
| `pangle_block_list_get` | Query | `/open_api/v1.3/pangle_block_list/get/` |
| `pangle_block_list_update` | Not implemented | — |
| `payment_portfolio_advertiser_get` | Not implemented | — |
| `payment_portfolio_advertiser_update` | Not implemented | — |
| `payment_portfolio_create` | Not implemented | — |
| `payment_portfolio_credit_line_update` | Not implemented | — |
| `payment_portfolio_get` | Not implemented | — |
| `payment_portfolio_user_get` | Not implemented | — |
| `pixel_create` | Not implemented | — |
| `pixel_event_create` | Not implemented | — |
| `pixel_event_delete` | Not implemented | — |
| `pixel_event_update` | Not implemented | — |
| `pixel_instant_page_event_get` | Not implemented | — |
| `pixel_list_get` | Query | `/open_api/v1.3/pixel/list/` |
| `pixel_update` | Not implemented | — |
| `playable_get` | Query | `/open_api/v1.3/playable/get/` |
| `playable_save` | Not implemented | — |
| `playable_validate` | Query | `/open_api/v1.3/playable/validate/` |
| `report_ad_benchmark_get` | Not implemented | — |
| `report_integrated_get` | Query | `/open_api/v1.3/report/integrated/get/` |
| `report_task_cancel` | Not implemented | — |
| `report_task_check` | Query | `/open_api/v1.3/report/task/check/` |
| `report_task_create` | Not implemented | — |
| `report_task_download` | Not implemented | — |
| `report_video_performance_get` | Not implemented | — |
| `rf_contract_query_get` | Not implemented | — |
| `rf_delivery_timezone_get` | Not implemented | — |
| `rf_inventory_estimate` | Not implemented | — |
| `rf_order_cancel` | Not implemented | — |
| `search_ad_negative_keyword_add` | Not implemented | — |
| `search_ad_negative_keyword_delete` | Not implemented | — |
| `search_ad_negative_keyword_get` | Not implemented | — |
| `search_ad_negative_keyword_update` | Not implemented | — |
| `search_region_get` | Query | `/open_api/v1.3/search/region/` |
| `showcase_identity_get` | Not implemented | — |
| `showcase_product_get` | Not implemented | — |
| `showcase_region_get` | Not implemented | — |
| `smart_plus_ad_appeal` | Not implemented | — |
| `smart_plus_ad_create` | Resource | — |
| `smart_plus_ad_get` | Resource, Query | `/open_api/v1.3/smart_plus/ad/get/` |
| `smart_plus_ad_material_status_update` | Not implemented | — |
| `smart_plus_ad_preview` | Not implemented | — |
| `smart_plus_ad_review_info_get` | Query | `/open_api/v1.3/smart_plus/ad/review_info/` |
| `smart_plus_ad_status_update` | Resource | — |
| `smart_plus_ad_update` | Resource | — |
| `smart_plus_adgroup_budget_update` | Not implemented | — |
| `smart_plus_adgroup_create` | Resource | — |
| `smart_plus_adgroup_get` | Resource, Query | `/open_api/v1.3/smart_plus/adgroup/get/` |
| `smart_plus_adgroup_status_update` | Resource | — |
| `smart_plus_adgroup_update` | Resource | — |
| `smart_plus_campaign_create` | Resource | — |
| `smart_plus_campaign_get` | Resource, Query | `/open_api/v1.3/smart_plus/campaign/get/` |
| `smart_plus_campaign_status_update` | Resource | — |
| `smart_plus_campaign_update` | Resource | — |
| `smart_plus_material_report_breakdown_run` | Not implemented | — |
| `smart_plus_material_report_overview_run` | Not implemented | — |
| `smart_plus_material_review_info_get` | Query | `/open_api/v1.3/smart_plus/material/review_info/` |
| `spark_ad_recommend_get` | Not implemented | — |
| `split_test_create` | Not implemented | — |
| `split_test_end_get` | Not implemented | — |
| `split_test_promote_run` | Not implemented | — |
| `split_test_result_get` | Not implemented | — |
| `split_test_update` | Not implemented | — |
| `store_list_get` | Not implemented | — |
| `store_product_get` | Not implemented | — |
| `subscription_get` | Not implemented | — |
| `subscription_subscribe_create` | Not implemented | — |
| `subscription_unsubscribe_cancel` | Not implemented | — |
| `targeting_search` | Not implemented | — |
| `tcm_tt_video_apply` | Not implemented | — |
| `tcm_tt_video_status_get` | Not implemented | — |
| `term_check` | Query | `/open_api/v1.3/term/check/` |
| `term_confirm` | Not implemented | — |
| `term_get` | Query | `/open_api/v1.3/term/get/` |
| `tiktok_inventory_filters_get` | Not implemented | — |
| `tiktok_inventory_filters_update` | Not implemented | — |
| `tool_action_category_get` | Query | `/open_api/v1.3/tool/action_category/` |
| `tool_bid_recommend` | Not implemented | — |
| `tool_brand_safety_partner_authorize_status_get` | Query | `/open_api/v1.3/tool/brand_safety/partner/authorize/status/` |
| `tool_carrier_get` | Query | `/open_api/v1.3/tool/carrier/` |
| `tool_content_exclusion_get` | Not implemented | — |
| `tool_content_exclusion_info_get` | Not implemented | — |
| `tool_contextual_tag_get` | Query | `/open_api/v1.3/tool/contextual_tag/get/` |
| `tool_contextual_tag_info_get` | Query | `/open_api/v1.3/tool/contextual_tag/info/` |
| `tool_device_model_get` | Query | `/open_api/v1.3/tool/device_model/` |
| `tool_diagnosis_get` | Not implemented | — |
| `tool_hashtag_get` | Query | `/open_api/v1.3/tool/hashtag/get/` |
| `tool_hashtag_recommend_search` | Not implemented | — |
| `tool_interest_category_get` | Query | `/open_api/v1.3/tool/interest_category/` |
| `tool_interest_keyword_get` | Query | `/open_api/v1.3/tool/interest_keyword/get/` |
| `tool_interest_keyword_recommend_search` | Not implemented | — |
| `tool_language_get` | Query | `/open_api/v1.3/tool/language/` |
| `tool_open_url_get` | Query | `/open_api/v1.3/tool/open_url/` |
| `tool_os_version_get` | Query | `/open_api/v1.3/tool/os_version/` |
| `tool_phone_region_code_get` | Query | `/open_api/v1.3/tool/phone_region_code/` |
| `tool_region_get` | Query | `/open_api/v1.3/tool/region/` |
| `tool_search_diagnosis_health_get` | Not implemented | — |
| `tool_search_keyword_idea_get` | Not implemented | — |
| `tool_search_keyword_recommend` | Not implemented | — |
| `tool_targeting_category_recommend_get` | Not implemented | — |
| `tool_targeting_info_get` | Not implemented | — |
| `tool_targeting_list_get` | Query | `/open_api/v1.3/tool/targeting/list/` |
| `tool_targeting_search` | Not implemented | — |
| `tool_timezone_get` | Query | `/open_api/v1.3/tool/timezone/` |
| `tool_url_validate` | Query | `/open_api/v1.3/tool/url_validate/` |
| `tool_vbo_status_check` | Not implemented | — |
| `tt_video_authorize_apply` | Not implemented | — |
| `tt_video_info_get` | Not implemented | — |
| `tt_video_list_get` | Not implemented | — |
| `tt_video_unbind` | Not implemented | — |
| `tto_oauth2_info_get` | Not implemented | — |
| `tto_oauth2_tcm_get` | Not implemented | — |
| `tto_tcm_anchor_create` | Not implemented | — |
| `tto_tcm_anchor_delete` | Not implemented | — |
| `tto_tcm_anchor_get` | Not implemented | — |
| `tto_tcm_brand_profile_create` | Not implemented | — |
| `tto_tcm_brand_profile_get` | Not implemented | — |
| `tto_tcm_campaign_create` | Not implemented | — |
| `tto_tcm_campaign_get` | Not implemented | — |
| `tto_tcm_campaign_link` | Not implemented | — |
| `tto_tcm_campaign_link_status_get` | Not implemented | — |
| `tto_tcm_campaign_update` | Not implemented | — |
| `tto_tcm_category_label_get` | Not implemented | — |
| `tto_tcm_rank_get` | Not implemented | — |
| `tto_tcm_report` | Not implemented | — |
| `user_info_get` | Not implemented | — |
| `video_fix_task_create` | Not implemented | — |
| `video_fix_task_get` | Not implemented | — |
