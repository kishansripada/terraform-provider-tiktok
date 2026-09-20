package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Source: spec/operations.json, official MCP schemas.
//
//go:embed campaigns.json
var specification []byte

type property struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Enum        []string  `json:"enum"`
	Items       *property `json:"items"`
}
type operation struct {
	Properties map[string]property `json:"properties"`
	Required   []string            `json:"required"`
}

var specs = func() map[string]operation {
	var s map[string]operation
	if err := json.Unmarshal(specification, &s); err != nil {
		panic(err)
	}
	return s
}()

func properties(kind string) map[string]property {
	result := map[string]property{}
	for _, suffix := range []string{"_create", "_update"} {
		for k, p := range specs[kind+suffix].Properties {
			result[k] = p
		}
	}
	for _, key := range []string{"advertiser_id", "request_id", "campaign_id"} {
		delete(result, key)
	}
	return result
}
func campaignSchema(kind string) schema.Schema {
	attrs := map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"advertiser_id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	for k, p := range properties(kind) {
		switch p.Type {
		case "string":
			attrs[k] = schema.StringAttribute{Optional: true, Description: p.Description}
		case "boolean":
			attrs[k] = schema.BoolAttribute{Optional: true, Description: p.Description}
		case "number":
			attrs[k] = schema.NumberAttribute{Optional: true, Description: p.Description}
		case "array":
			attrs[k] = schema.ListAttribute{Optional: true, ElementType: types.StringType, Description: p.Description}
		default:
			panic("unsupported campaign property: " + k)
		}
	}
	return schema.Schema{Description: "A TikTok campaign. Omitted fields are unmanaged. Missing/deleted campaigns and immutable changes fail closed; replacement is never automatic.", Attributes: attrs}
}
func normalize(kind, key string, value any) any {
	if properties(kind)[key].Type == "number" {
		if s, ok := value.(string); ok {
			if n, err := strconv.ParseFloat(s, 64); err == nil {
				return n
			}
		}
	}
	return value
}
func validateValue(p property, value any) error {
	valid := false
	switch p.Type {
	case "string":
		_, valid = value.(string)
	case "boolean":
		_, valid = value.(bool)
	case "number":
		_, valid = value.(float64)
	case "array":
		items, ok := value.([]any)
		valid = ok
		for _, item := range items {
			if p.Items == nil || validateValue(*p.Items, item) != nil {
				valid = false
			}
		}
	}
	if !valid {
		return fmt.Errorf("expected %s", p.Type)
	}
	if len(p.Enum) > 0 {
		for _, option := range p.Enum {
			if value == option {
				return nil
			}
		}
		return fmt.Errorf("value must be one of %v", p.Enum)
	}
	return nil
}
func validateAttributes(kind string, attrs map[string]any) error {
	props := properties(kind)
	for k, v := range attrs {
		p, ok := props[k]
		if !ok {
			return fmt.Errorf("unknown/read-only attribute %s", k)
		}
		if err := validateValue(p, v); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
	}
	return nil
}
func validateCreate(kind string, attrs map[string]any) error {
	if err := validateAttributes(kind, attrs); err != nil {
		return err
	}
	for _, key := range specs[kind+"_create"].Required {
		if key == "advertiser_id" || key == "request_id" {
			continue
		}
		if _, ok := attrs[key]; !ok {
			return fmt.Errorf("%s is required for creation", key)
		}
	}
	if attrs["operation_status"] != "DISABLE" {
		return fmt.Errorf("new campaigns must explicitly set operation_status = DISABLE; enable in a later apply")
	}
	if attrs["budget_mode"] == nil {
		return fmt.Errorf("budget_mode must be explicit")
	}
	if attrs["budget_mode"] != "BUDGET_MODE_INFINITE" {
		if n, _ := attrs["budget"].(float64); n <= 0 {
			return fmt.Errorf("positive budget is required")
		}
	}
	if attrs["objective_type"] == "RF_REACH" {
		return fmt.Errorf("Reach & Frequency creation is unsupported")
	}
	if attrs["objective_type"] == "APP_PROMOTION" && attrs["app_promotion_type"] == nil {
		return fmt.Errorf("APP_PROMOTION requires app_promotion_type")
	}
	if kind == "smart_plus_campaign" && attrs["objective_type"] == "WEB_CONVERSIONS" && attrs["sales_destination"] == nil {
		return fmt.Errorf("WEB_CONVERSIONS requires sales_destination")
	}
	if attrs["campaign_type"] == "IOS14_CAMPAIGN" && attrs["app_id"] == nil {
		return fmt.Errorf("Dedicated campaign requires app_id")
	}
	if attrs["postback_window_mode"] != nil && attrs["campaign_type"] != "IOS14_CAMPAIGN" {
		return fmt.Errorf("postback_window_mode requires a disabled Dedicated Campaign")
	}
	return nil
}
func delta(kind string, before, after map[string]any) map[string]any {
	d := map[string]any{}
	for k, v := range after {
		if !reflect.DeepEqual(normalize(kind, k, before[k]), v) {
			d[k] = v
		}
	}
	return d
}
func validateUpdate(kind string, current, changed map[string]any) error {
	for k, v := range changed {
		if k == "operation_status" {
			continue
		}
		if _, ok := specs[kind+"_update"].Properties[k]; !ok {
			return fmt.Errorf("immutable attribute %s: create a separate campaign explicitly", k)
		}
		if k == "budget" {
			if n, _ := v.(float64); n <= 0 {
				return fmt.Errorf("budget must be positive")
			}
		}
		if k == "budget_auto_adjust_strategy" && (current["budget_optimize_on"] != true || current["budget_mode"] != "BUDGET_MODE_DYNAMIC_DAILY_BUDGET" || current["app_promotion_type"] == "MINIS") {
			return fmt.Errorf("budget auto-adjust requires dynamic daily CBO and non-MINIS")
		}
		if k == "special_industries" {
			old, ok := current[k].([]any)
			if !ok {
				return fmt.Errorf("special industries may only be removed")
			}
			for _, x := range v.([]any) {
				found := false
				for _, y := range old {
					if x == y {
						found = true
					}
				}
				if !found {
					return fmt.Errorf("special industries may only be removed")
				}
			}
		}
	}
	if changed["operation_status"] == "DISABLE" && current["campaign_type"] == "IOS14_CAMPAIGN" && current["postback_window_mode"] == nil {
		return fmt.Errorf("disabling this Dedicated Campaign may implicitly set a SKAN postback window; perform that specialized transition explicitly outside Terraform, then refresh")
	}
	return nil
}

// Encode/decode through Terraform's typed JSON codec so IDs remain strings and
// null (unmanaged) stays distinct from false, zero, and empty lists.
func rawMap(raw tftypes.Value) (map[string]any, error) {
	values := map[string]tftypes.Value{}
	if raw.IsNull() {
		return map[string]any{}, nil
	}
	if err := raw.As(&values); err != nil {
		return nil, err
	}
	result := map[string]any{}
	for k, v := range values {
		if v.IsNull() || !v.IsFullyKnown() {
			continue
		}
		d, err := valueGo(v)
		if err != nil {
			return nil, err
		}
		result[k] = d
	}
	return result, nil
}
func valueGo(v tftypes.Value) (any, error) {
	switch {
	case v.Type().Is(tftypes.String):
		var s string
		err := v.As(&s)
		return s, err
	case v.Type().Is(tftypes.Bool):
		var b bool
		err := v.As(&b)
		return b, err
	case v.Type().Is(tftypes.Number):
		var n big.Float
		err := v.As(&n)
		value, _ := n.Float64()
		return value, err
	default:
		var list []tftypes.Value
		if err := v.As(&list); err != nil {
			return nil, err
		}
		result := []any{}
		for _, item := range list {
			x, err := valueGo(item)
			if err != nil {
				return nil, err
			}
			result = append(result, x)
		}
		return result, nil
	}
}
func managed(kind string, values map[string]any) map[string]any {
	result := map[string]any{}
	for k := range properties(kind) {
		if v, ok := values[k]; ok {
			result[k] = v
		}
	}
	return result
}
