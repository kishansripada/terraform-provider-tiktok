package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//go:embed objects.json
var objectSpecification []byte
var objectSpecs = func() map[string]operation {
	var result map[string]operation
	if err := json.Unmarshal(objectSpecification, &result); err != nil {
		panic(err)
	}
	return result
}()

func objectID(kind string) string {
	if kind == "smart_plus_ad" {
		return "smart_plus_ad_id"
	}
	return strings.TrimPrefix(kind, "smart_plus_") + "_id"
}
func objectOperation(kind, suffix string) operation {
	op := objectSpecs[kind+suffix]
	if kind == "ad" && (suffix == "_create" || suffix == "_update") {
		creative := op.Properties["creatives"].Items
		props := map[string]property{"adgroup_id": {Type: "string"}}
		for k, v := range creative.Properties {
			props[k] = v
		}
		return operation{Properties: props, Required: append([]string{"adgroup_id"}, creative.Required...)}
	}
	return op
}
func objectProperties(kind string) map[string]property {
	props := map[string]property{}
	for _, suffix := range []string{"_create", "_update"} {
		for k, v := range objectOperation(kind, suffix).Properties {
			if old, ok := props[k]; ok {
				props[k] = mergeProperty(old, v)
			} else {
				props[k] = v
			}
		}
	}
	for _, k := range []string{"advertiser_id", "request_id", objectID(kind)} {
		delete(props, k)
	}
	return props
}

// Preserve create-only nested fields while exposing update-only fields as well.
func mergeProperty(a, b property) property {
	if a.Type == "object" && b.Type == "object" {
		props := map[string]property{}
		for k, v := range a.Properties {
			props[k] = v
		}
		for k, v := range b.Properties {
			if old, ok := props[k]; ok {
				props[k] = mergeProperty(old, v)
			} else {
				props[k] = v
			}
		}
		b.Properties = props
		b.Required = nil // Operation-specific requirements are checked separately.
	}
	if a.Type == "array" && b.Type == "array" && a.Items != nil && b.Items != nil {
		item := mergeProperty(*a.Items, *b.Items)
		b.Items = &item
	}
	return b
}

// API responses include output-only metadata. Only documented input keys may be
// carried into replacement requests; retain nested values outside HCL ownership.
func writableValue(p property, value any) any {
	if p.Type == "object" {
		input, ok := value.(map[string]any)
		if !ok {
			return value
		}
		out := map[string]any{}
		for k, child := range p.Properties {
			if v := input[k]; v != nil {
				out[k] = writableValue(child, v)
			}
		}
		return out
	}
	if p.Type == "array" {
		input, ok := value.([]any)
		if !ok {
			return value
		}
		out := make([]any, len(input))
		for i, v := range input {
			out[i] = writableValue(*p.Items, v)
		}
		return out
	}
	return typedObjectValue(p, value)
}
func propertyType(p property) attr.Type {
	switch p.Type {
	case "string":
		return types.StringType
	case "number", "integer":
		return types.NumberType
	case "boolean":
		return types.BoolType
	case "object":
		fields := map[string]attr.Type{}
		for k, v := range p.Properties {
			fields[k] = propertyType(v)
		}
		return types.ObjectType{AttrTypes: fields}
	case "array":
		return types.ListType{ElemType: propertyType(*p.Items)}
	default:
		panic("unsupported schema type: " + p.Type)
	}
}
func propertyAttribute(p property) schema.Attribute {
	switch p.Type {
	case "string":
		return schema.StringAttribute{Optional: true, Description: p.Description}
	case "number", "integer":
		return schema.NumberAttribute{Optional: true, Description: p.Description}
	case "boolean":
		return schema.BoolAttribute{Optional: true, Description: p.Description}
	case "object":
		fields := map[string]schema.Attribute{}
		for k, v := range p.Properties {
			fields[k] = propertyAttribute(v)
		}
		return schema.SingleNestedAttribute{Optional: true, Description: p.Description, Attributes: fields}
	case "array":
		if p.Items.Type == "object" {
			fields := map[string]schema.Attribute{}
			for k, v := range p.Items.Properties {
				fields[k] = propertyAttribute(v)
			}
			return schema.ListNestedAttribute{Optional: true, Description: p.Description, NestedObject: schema.NestedAttributeObject{Attributes: fields}}
		}
		return schema.ListAttribute{Optional: true, Description: p.Description, ElementType: propertyType(*p.Items)}
	default:
		panic("unsupported schema type: " + p.Type)
	}
}
func objectSchema(kind string) schema.Schema {
	props := map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"advertiser_id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"observed_json": schema.StringAttribute{Computed: true, Description: "Last observed API object, including server-generated metadata."},
	}
	for k, p := range objectProperties(kind) {
		props[k] = propertyAttribute(p)
	}
	return schema.Schema{Description: "A single TikTok " + kind + ". Omitted fields are unmanaged; new resources must be disabled. Replacement is never automatic.", Attributes: props}
}
func objectManaged(kind string, values map[string]any) map[string]any {
	result := map[string]any{}
	for k := range objectProperties(kind) {
		if v, ok := values[k]; ok && v != nil {
			result[k] = v
		}
	}
	return result
}
func validateObject(kind string, values map[string]any, creating bool) error {
	for k, v := range values {
		p, ok := objectProperties(kind)[k]
		if !ok {
			return fmt.Errorf("unknown attribute %s", k)
		}
		if err := validateValue(p, v); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
	}
	if status, ok := values["operation_status"]; ok && status != "ENABLE" && status != "DISABLE" {
		return fmt.Errorf("operation_status must be ENABLE or DISABLE")
	}
	if budget, ok := values["budget"].(float64); ok && budget <= 0 {
		return fmt.Errorf("budget must be positive")
	}
	if creating && values["schedule_type"] == "SCHEDULE_START_END" && values["schedule_end_time"] == nil {
		return fmt.Errorf("schedule_end_time is required for SCHEDULE_START_END")
	}
	if creating {
		for k, v := range values {
			p, ok := objectOperation(kind, "_create").Properties[k]
			if !ok {
				return fmt.Errorf("%s can only be set after creation", k)
			}
			if err := validateValue(p, v); err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
		}
		if values["operation_status"] != "DISABLE" {
			return fmt.Errorf("create disabled first; enable in a separate apply")
		}
		for _, k := range objectOperation(kind, "_create").Required {
			if k == "advertiser_id" || k == "request_id" {
				continue
			}
			if values[k] == nil {
				return fmt.Errorf("%s is required for creation", k)
			}
		}
	}
	for _, key := range []string{"campaign_id", "adgroup_id"} {
		if v, ok := values[key]; ok && !validID(fmt.Sprint(v)) {
			return fmt.Errorf("invalid %s", key)
		}
	}
	return nil
}

// Terraform object types require every key, including null optional children.
func typedObjectValue(p property, value any) any {
	if value == nil {
		return nil
	}
	switch p.Type {
	case "object":
		input, ok := value.(map[string]any)
		if !ok {
			return value
		}
		result := map[string]any{}
		for k, v := range p.Properties {
			result[k] = typedObjectValue(v, input[k])
		}
		return result
	case "array":
		input, ok := value.([]any)
		if !ok {
			return value
		}
		result := make([]any, len(input))
		for i, v := range input {
			result[i] = typedObjectValue(*p.Items, v)
		}
		return result
	case "number", "integer":
		if s, ok := value.(string); ok {
			var n json.Number = json.Number(s)
			if f, err := n.Float64(); err == nil {
				return f
			}
		}
	}
	return value
}

// Project only fields the user manages. API-generated nested metadata must not
// become desired configuration or overwrite optional nulls in Terraform's plan.
func observedProjection(desired, observed any) any {
	if d, ok := desired.(map[string]any); ok {
		o, _ := observed.(map[string]any)
		result := map[string]any{}
		for k, v := range d {
			result[k] = observedProjection(v, o[k])
		}
		return result
	}
	if d, ok := desired.([]any); ok {
		o, ok := observed.([]any)
		if !ok || len(d) != len(o) {
			return observed
		}
		result := make([]any, len(d))
		for i, v := range d {
			result[i] = observedProjection(v, o[i])
		}
		return result
	}
	return observed
}
func objectDelta(kind string, current, desired map[string]any) map[string]any {
	result := map[string]any{}
	for k, v := range desired {
		p := objectProperties(kind)[k]
		if !reflect.DeepEqual(typedObjectValue(p, v), typedObjectValue(p, observedProjection(v, current[k]))) {
			result[k] = v
		}
	}
	return result
}
func (c *apiClient) readObject(ctx context.Context, kind, id string) (map[string]any, error) {
	if !validID(id) {
		return nil, fmt.Errorf("invalid or pending identity; import the verified remote object before retrying")
	}
	rows, err := c.listFiltered(ctx, kind, map[string]any{"primary_status": "STATUS_ALL", objectID(kind) + "s": []string{id}})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row[objectID(kind)] == id {
			return row, nil
		}
	}
	return nil, fmt.Errorf("managed %s %s missing; refusing automatic recreation", kind, id)
}
func (c *apiClient) verifyObject(ctx context.Context, kind, id string, desired map[string]any, deleting bool) (map[string]any, error) {
	var last map[string]any
	for i := 0; i < 4; i++ {
		row, err := c.readObject(ctx, kind, id)
		if err != nil {
			return last, err
		}
		last = row
		if (deleting && deleted(row)) || (!deleting && !deleted(row) && len(objectDelta(kind, row, desired)) == 0) {
			return row, nil
		}
		if i < 3 {
			select {
			case <-ctx.Done():
				return last, ctx.Err()
			case <-time.After(time.Duration(250*(1<<i)) * time.Millisecond):
			}
		}
	}
	return last, fmt.Errorf("remote verification did not converge for %s %s", kind, id)
}

type objectResource struct {
	kind   string
	client *apiClient
}

func (r *objectResource) Metadata(_ context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_" + r.kind
}
func (r *objectResource) Schema(_ context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = objectSchema(r.kind)
}
func (r *objectResource) Configure(_ context.Context, req resource.ConfigureRequest, res *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	r.client, ok = req.ProviderData.(*apiClient)
	if !ok {
		addError(&res.Diagnostics, fmt.Errorf("unexpected provider data"))
	}
}
func (r *objectResource) identity(raw tftypes.Value) (string, map[string]any, error) {
	values, err := rawMap(raw)
	if err != nil {
		return "", nil, err
	}
	if r.client == nil || values["advertiser_id"] != r.client.account {
		return "", nil, fmt.Errorf("state advertiser mismatch")
	}
	id, _ := values["id"].(string)
	if !validID(id) {
		return "", nil, fmt.Errorf("pending or invalid identity; reconcile and import the verified object")
	}
	return id, objectManaged(r.kind, values), nil
}
func (r *objectResource) save(state *tfsdk.State, id string, desired, observed map[string]any) error {
	values := map[string]any{"id": id, "advertiser_id": r.client.account, "observed_json": nil}
	if observed != nil {
		b, err := json.Marshal(observed)
		if err != nil {
			return err
		}
		values["observed_json"] = string(b)
	}
	for k, p := range objectProperties(r.kind) {
		v := desired[k]
		if observed != nil && v != nil {
			v = observedProjection(v, observed[k])
		}
		values[k] = typedObjectValue(p, v)
	}
	b, err := json.Marshal(values)
	if err != nil {
		return err
	}
	raw, err := tftypes.ValueFromJSON(b, state.Schema.Type().TerraformType(context.Background()))
	if err == nil {
		state.Raw = raw
	}
	return err
}
func (r *objectResource) ValidateConfig(_ context.Context, req resource.ValidateConfigRequest, res *resource.ValidateConfigResponse) {
	values, err := rawMap(req.Config.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	addError(&res.Diagnostics, validateObject(r.kind, objectManaged(r.kind, values), false))
}
func (r *objectResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		if r.client != nil && !r.client.allowDelete {
			addError(&res.Diagnostics, fmt.Errorf("deletion disabled"))
		}
		return
	}
	// Unknown dependency references resolve at apply; don't misdiagnose them as missing.
	if !req.Config.Raw.IsFullyKnown() {
		return
	}
	values, err := rawMap(req.Plan.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	desired := objectManaged(r.kind, values)
	if req.State.Raw.IsNull() {
		addError(&res.Diagnostics, validateObject(r.kind, desired, true))
		return
	}
	prior, err := rawMap(req.State.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	for k := range objectDelta(r.kind, prior, desired) {
		if k == "operation_status" {
			continue
		}
		if _, ok := objectOperation(r.kind, "_update").Properties[k]; !ok || k == "campaign_id" || k == "adgroup_id" {
			addError(&res.Diagnostics, fmt.Errorf("immutable attribute %s; create a separate resource explicitly", k))
		}
	}
}
func (r *objectResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	id, desired, err := r.identity(req.State.Raw)
	if err == nil {
		var row map[string]any
		row, err = r.client.readObject(ctx, r.kind, id)
		if err == nil && deleted(row) {
			err = fmt.Errorf("managed resource is deleted; refusing automatic recreation")
		}
		if err == nil {
			err = r.save(&res.State, id, desired, row)
		}
	}
	addError(&res.Diagnostics, err)
}
func (r *objectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] != r.client.account || !validID(parts[1]) {
		addError(&res.Diagnostics, fmt.Errorf("import requires advertiser_id/resource_id matching the provider"))
		return
	}
	row, err := r.client.readObject(ctx, r.kind, parts[1])
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if deleted(row) {
		addError(&res.Diagnostics, fmt.Errorf("cannot import deleted resource"))
		return
	}
	desired := map[string]any{}
	for k, p := range objectProperties(r.kind) {
		v := writableValue(p, row[k])
		if v != nil && validateValue(p, v) == nil {
			desired[k] = v
		}
	}
	addError(&res.Diagnostics, r.save(&res.State, parts[1], desired, row))
}
func (r *objectResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	values, err := rawMap(req.Plan.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	desired := objectManaged(r.kind, values)
	if err = validateObject(r.kind, desired, true); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if !r.client.writes {
		addError(&res.Diagnostics, fmt.Errorf("writes disabled"))
		return
	}
	params := map[string]any{"advertiser_id": r.client.account}
	if r.kind == "ad" {
		creative := map[string]any{}
		for k, v := range desired {
			if k != "adgroup_id" {
				creative[k] = v
			}
		}
		params["adgroup_id"] = desired["adgroup_id"]
		params["creatives"] = []any{creative}
	} else {
		for k, v := range desired {
			params[k] = v
		}
	}
	requestID := fmt.Sprint(time.Now().UnixNano())
	if _, ok := objectSpecs[r.kind+"_create"].Properties["request_id"]; ok {
		params["request_id"] = requestID
	}
	if err = r.save(&res.State, "pending-"+requestID, desired, nil); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	response, err := r.client.call(ctx, r.kind+"_create", params)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	id, _ := response[objectID(r.kind)].(string)
	if r.kind == "ad" {
		ids, ok := response["ad_ids"].([]any)
		if ok && len(ids) == 1 {
			id, _ = ids[0].(string)
		}
	}
	if !validID(id) {
		addError(&res.Diagnostics, fmt.Errorf("create returned no unique valid ID; reconcile pending request %s", requestID))
		return
	}
	if err = r.save(&res.State, id, desired, nil); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	row, err := r.client.verifyObject(ctx, r.kind, id, desired, false)
	if row != nil {
		addError(&res.Diagnostics, r.save(&res.State, id, desired, row))
	}
	addError(&res.Diagnostics, err)
}

func mergeObject(base, override any) any {
	b, bok := base.(map[string]any)
	o, ook := override.(map[string]any)
	if !bok || !ook {
		return override
	}
	result := map[string]any{}
	for k, v := range b {
		result[k] = v
	}
	for k, v := range o {
		result[k] = mergeObject(b[k], v)
	}
	return result
}
func reuseMaterials(current, desired any) any {
	live, _ := current.([]any)
	wanted, ok := desired.([]any)
	if !ok {
		return desired
	}
	result := make([]any, len(wanted))
	used := map[int]bool{}
	for i, item := range wanted {
		fields, ok := item.(map[string]any)
		if !ok {
			result[i] = item
			continue
		}
		copy := map[string]any{}
		for k, v := range fields {
			copy[k] = v
		}
		if copy["ad_material_id"] == nil && copy["creative_info"] != nil {
			for j, candidate := range live {
				row, ok := candidate.(map[string]any)
				if !ok || used[j] || row["ad_material_id"] == nil {
					continue
				}
				if reflect.DeepEqual(copy["creative_info"], observedProjection(copy["creative_info"], row["creative_info"])) {
					copy["ad_material_id"] = row["ad_material_id"]
					used[j] = true
					break
				}
			}
		}
		result[i] = copy
	}
	return result
}
func (r *objectResource) updateParams(id string, current, changed map[string]any) (map[string]any, error) {
	props := objectOperation(r.kind, "_update").Properties
	full := r.kind == "adgroup" || r.kind == "ad"
	fields := map[string]any{}
	if full {
		// These endpoints replace settings: carry forward every readable writable
		// field, including settings omitted from HCL, instead of sending a patch.
		for k, p := range props {
			if k == "advertiser_id" || k == objectID(r.kind) {
				continue
			}
			if v := writableValue(p, current[k]); v != nil {
				if err := validateValue(p, v); err != nil {
					return nil, fmt.Errorf("cannot preserve remote %s: %w", k, err)
				}
				fields[k] = v
			}
		}
	}
	for k, v := range changed {
		if k == "campaign_id" || k == "adgroup_id" {
			return nil, fmt.Errorf("parent identity cannot change")
		}
		if _, ok := props[k]; !ok {
			return nil, fmt.Errorf("immutable attribute %s", k)
		}
		// Smart+ nested object updates may replace whole product_info objects.
		// Creative lists are replacements. Reuse the material identity for an
		// unchanged creative, including when the user reorders the list.
		if r.kind == "smart_plus_ad" && k == "creative_list" {
			v = reuseMaterials(current[k], v)
		}
		fields[k] = mergeObject(writableValue(props[k], current[k]), v)
		if err := validateValue(props[k], fields[k]); err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
	}
	params := map[string]any{"advertiser_id": r.client.account, objectID(r.kind): id}
	if r.kind == "ad" {
		if !validID(fmt.Sprint(current["adgroup_id"])) {
			return nil, fmt.Errorf("cannot identify ad's parent")
		}
		delete(fields, "adgroup_id")
		fields["ad_id"] = id
		delete(params, "ad_id")
		params["adgroup_id"] = current["adgroup_id"]
		params["creatives"] = []any{fields}
	} else {
		for k, v := range fields {
			params[k] = v
		}
	}
	return params, nil
}
func (r *objectResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	id, prior, err := r.identity(req.State.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	values, err := rawMap(req.Plan.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	desired := objectManaged(r.kind, values)
	if err = validateObject(r.kind, desired, false); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	current, err := r.client.readObject(ctx, r.kind, id)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if deleted(current) {
		addError(&res.Diagnostics, fmt.Errorf("resource deleted"))
		return
	}
	if len(objectDelta(r.kind, current, prior)) > 0 {
		addError(&res.Diagnostics, fmt.Errorf("stale plan: remote managed values changed; replan"))
		return
	}
	changed := objectDelta(r.kind, current, desired)
	status, statusChanged := changed["operation_status"]
	delete(changed, "operation_status")
	if r.kind == "smart_plus_adgroup" {
		for _, field := range []string{"bid_price", "conversion_bid_price", "roas_bid"} {
			if _, ok := changed[field]; !ok {
				continue
			}
			siblings, e := r.client.list(ctx, r.kind, fmt.Sprint(current["campaign_id"]))
			if e != nil {
				addError(&res.Diagnostics, e)
				return
			}
			for _, s := range siblings {
				if s["campaign_id"] == current["campaign_id"] && s["adgroup_id"] != id && !deleted(s) {
					addError(&res.Diagnostics, fmt.Errorf("%s changes sibling ad groups; manage this campaign-wide transition explicitly", field))
					return
				}
			}
		}
	}
	type step struct {
		operation        string
		params, expected map[string]any
	}
	steps := []step{}
	if len(changed) > 0 {
		params, e := r.updateParams(id, current, changed)
		if e != nil {
			addError(&res.Diagnostics, e)
			return
		}
		steps = append(steps, step{r.kind + "_update", params, changed})
	}
	if statusChanged {
		s := step{r.kind + "_status_update", map[string]any{"advertiser_id": r.client.account, objectID(r.kind) + "s": []string{id}, "operation_status": status}, map[string]any{"operation_status": status}}
		if status == "DISABLE" {
			steps = append([]step{s}, steps...)
		} else {
			steps = append(steps, s)
		}
	}
	res.State = req.State
	for _, s := range steps {
		_, err = r.client.call(ctx, s.operation, s.params)
		if err != nil {
			addError(&res.Diagnostics, err)
			return
		}
		current, err = r.client.verifyObject(ctx, r.kind, id, s.expected, false)
		if current != nil {
			addError(&res.Diagnostics, r.save(&res.State, id, desired, current))
		}
		if err != nil || res.Diagnostics.HasError() {
			addError(&res.Diagnostics, err)
			return
		}
	}
	current, err = r.client.verifyObject(ctx, r.kind, id, desired, false)
	if current != nil {
		addError(&res.Diagnostics, r.save(&res.State, id, desired, current))
	}
	addError(&res.Diagnostics, err)
}
func (r *objectResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	id, _, err := r.identity(req.State.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if !r.client.writes || !r.client.allowDelete {
		addError(&res.Diagnostics, fmt.Errorf("deletion requires allow_writes and allow_delete"))
		return
	}
	row, err := r.client.readObject(ctx, r.kind, id)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if !deleted(row) {
		if strings.HasSuffix(r.kind, "adgroup") {
			for _, kind := range []string{"ad", "smart_plus_ad"} {
				children, e := r.client.listFiltered(ctx, kind, map[string]any{"primary_status": "STATUS_ALL", "adgroup_ids": []string{id}})
				if e != nil {
					addError(&res.Diagnostics, e)
					return
				}
				for _, child := range children {
					if !deleted(child) && !validID(fmt.Sprint(child["adgroup_id"])) {
						addError(&res.Diagnostics, fmt.Errorf("child response omits its parent identity; deletion blocked"))
						return
					}
					if child["adgroup_id"] == id && !deleted(child) {
						addError(&res.Diagnostics, fmt.Errorf("ad group has nondeleted children"))
						return
					}
				}
			}
		}
		_, err = r.client.call(ctx, r.kind+"_status_update", map[string]any{"advertiser_id": r.client.account, objectID(r.kind) + "s": []string{id}, "operation_status": "DELETE"})
		if err != nil {
			addError(&res.Diagnostics, err)
			return
		}
		if _, err = r.client.verifyObject(ctx, r.kind, id, nil, true); err != nil {
			addError(&res.Diagnostics, err)
			return
		}
	}
	res.State.RemoveResource(ctx)
}
