package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type campaignResource struct {
	kind   string
	client *apiClient
}

func (r *campaignResource) Metadata(_ context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_" + r.kind
}
func (r *campaignResource) Schema(_ context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = campaignSchema(r.kind)
}
func (r *campaignResource) Configure(_ context.Context, req resource.ConfigureRequest, res *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	r.client, ok = req.ProviderData.(*apiClient)
	if !ok {
		res.Diagnostics.AddError("Provider error", "Unexpected provider data")
	}
}
func (r *campaignResource) ValidateConfig(_ context.Context, req resource.ValidateConfigRequest, res *resource.ValidateConfigResponse) {
	values, err := rawMap(req.Config.Raw)
	if err == nil {
		err = validateAttributes(r.kind, managed(r.kind, values))
	}
	addError(&res.Diagnostics, err)
}
func (r *campaignResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, res *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		if r.client != nil && !r.client.allowDelete {
			addError(&res.Diagnostics, fmt.Errorf("campaign deletion disabled in provider"))
		}
		return
	}
	plan, err := rawMap(req.Plan.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if req.State.Raw.IsNull() {
		// Unknown input references are checked again once resolved at apply.
		values := map[string]tftypes.Value{}
		if err := req.Config.Raw.As(&values); err != nil {
			addError(&res.Diagnostics, err)
			return
		}
		for _, v := range values {
			if !v.IsFullyKnown() {
				return
			}
		}
		addError(&res.Diagnostics, validateCreate(r.kind, managed(r.kind, plan)))
		return
	}
	prior, err := rawMap(req.State.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if r.client != nil && prior["advertiser_id"] != r.client.account {
		addError(&res.Diagnostics, fmt.Errorf("state advertiser does not match provider"))
		return
	}
	// Cross-field rules that need response-only metadata are checked against
	// the live campaign immediately before mutation.
	for field := range delta(r.kind, prior, managed(r.kind, plan)) {
		if field != "operation_status" {
			if _, ok := specs[r.kind+"_update"].Properties[field]; !ok {
				addError(&res.Diagnostics, fmt.Errorf("immutable attribute %s; replacement is disabled", field))
			}
		}
	}
}
func addError(d *diag.Diagnostics, err error) {
	if err != nil {
		d.AddError("TikTok operation blocked", err.Error())
	}
}

func (r *campaignResource) save(state *tfsdk.State, id string, desired, observed map[string]any) error {
	values := map[string]any{"id": id, "advertiser_id": r.client.account}
	for k := range properties(r.kind) {
		values[k] = nil
		if v, managed := desired[k]; managed {
			values[k] = v
			if observed != nil {
				values[k] = normalize(r.kind, k, observed[k])
			}
		}
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
func (r *campaignResource) identity(raw tftypes.Value) (string, map[string]any, error) {
	values, err := rawMap(raw)
	if err != nil {
		return "", nil, err
	}
	if r.client == nil {
		return "", nil, fmt.Errorf("provider not configured")
	}
	if values["advertiser_id"] != r.client.account {
		return "", nil, fmt.Errorf("state advertiser mismatch")
	}
	id, _ := values["id"].(string)
	if !validID(id) {
		return "", nil, fmt.Errorf("ambiguous or invalid campaign ID %q; inspect TikTok and import the verified ID before continuing", id)
	}
	return id, managed(r.kind, values), nil
}
func (r *campaignResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	id, attrs, err := r.identity(req.State.Raw)
	if err == nil {
		var row map[string]any
		row, err = r.client.read(ctx, r.kind, id)
		if err == nil && deleted(row) {
			err = fmt.Errorf("campaign %s is terminally deleted; refusing automatic recreation", id)
		}
		if err == nil {
			err = r.save(&res.State, id, attrs, row)
		}
	}
	addError(&res.Diagnostics, err)
}
func (r *campaignResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] != r.client.account || !validID(parts[1]) {
		addError(&res.Diagnostics, fmt.Errorf("import ID must be advertiser_id/campaign_id matching the provider"))
		return
	}
	row, err := r.client.read(ctx, r.kind, parts[1])
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if deleted(row) {
		addError(&res.Diagnostics, fmt.Errorf("cannot import terminally deleted campaign"))
		return
	}
	attrs := map[string]any{}
	for key, p := range properties(r.kind) {
		v := normalize(r.kind, key, row[key])
		if validateValue(p, v) == nil {
			attrs[key] = v
		}
	}
	addError(&res.Diagnostics, r.save(&res.State, parts[1], attrs, nil))
}
func (r *campaignResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	values, err := rawMap(req.Plan.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	attrs := managed(r.kind, values)
	if err = validateCreate(r.kind, attrs); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if !r.client.writes {
		addError(&res.Diagnostics, fmt.Errorf("writes disabled"))
		return
	}
	requestID := strconv.FormatInt(time.Now().UnixNano(), 10)
	// If the response is ambiguous, persist a blocking identity rather than
	// permitting a second create on the next apply. Import repairs identity.
	if err = r.save(&res.State, "pending-"+requestID, attrs, nil); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	params := map[string]any{"advertiser_id": r.client.account, "request_id": requestID}
	for k, v := range attrs {
		params[k] = v
	}
	result, err := r.client.call(ctx, r.kind+"_create", params)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	id, _ := result["campaign_id"].(string)
	if !validID(id) {
		addError(&res.Diagnostics, fmt.Errorf("create returned no valid campaign ID; reconcile pending request %s", requestID))
		return
	}
	if err = r.save(&res.State, id, attrs, nil); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	row, err := r.client.verify(ctx, r.kind, id, attrs, false)
	if row != nil {
		addError(&res.Diagnostics, r.save(&res.State, id, attrs, row))
	}
	addError(&res.Diagnostics, err)
}
func (r *campaignResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
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
	attrs := managed(r.kind, values)
	if err = validateAttributes(r.kind, attrs); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	current, err := r.client.read(ctx, r.kind, id)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if deleted(current) {
		addError(&res.Diagnostics, fmt.Errorf("campaign is deleted"))
		return
	}
	for k, v := range prior {
		if !reflect.DeepEqual(normalize(r.kind, k, current[k]), v) {
			addError(&res.Diagnostics, fmt.Errorf("stale plan: %s changed remotely; replan", k))
			return
		}
	}
	changed := delta(r.kind, current, attrs)
	if err = validateUpdate(r.kind, current, changed); err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	status, statusChanged := changed["operation_status"]
	delete(changed, "operation_status")
	type request struct {
		operation        string
		params, expected map[string]any
	}
	var requests []request
	if len(changed) > 0 {
		params := map[string]any{"advertiser_id": r.client.account, "campaign_id": id}
		for k, v := range changed {
			params[k] = v
		}
		requests = append(requests, request{r.kind + "_update", params, changed})
	}
	if statusChanged {
		s := request{r.kind + "_status_update", map[string]any{"advertiser_id": r.client.account, "campaign_ids": []string{id}, "operation_status": status}, map[string]any{"operation_status": status}}
		if status == "DISABLE" {
			requests = append([]request{s}, requests...)
		} else {
			requests = append(requests, s)
		}
	}
	res.State = req.State
	for _, request := range requests {
		_, err = r.client.call(ctx, request.operation, request.params)
		if err != nil {
			addError(&res.Diagnostics, err)
			return
		}
		current, err = r.client.verify(ctx, r.kind, id, request.expected, false)
		if current != nil {
			addError(&res.Diagnostics, r.save(&res.State, id, attrs, current))
		}
		if err != nil || res.Diagnostics.HasError() {
			addError(&res.Diagnostics, err)
			return
		}
	}
	current, err = r.client.verify(ctx, r.kind, id, attrs, false)
	if current != nil {
		addError(&res.Diagnostics, r.save(&res.State, id, attrs, current))
	}
	addError(&res.Diagnostics, err)
}
func (r *campaignResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	id, _, err := r.identity(req.State.Raw)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if !r.client.writes || !r.client.allowDelete {
		addError(&res.Diagnostics, fmt.Errorf("deletion requires provider allow_writes and allow_delete"))
		return
	}
	row, err := r.client.read(ctx, r.kind, id)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	if !deleted(row) {
		for _, kind := range []string{"adgroup", "smart_plus_adgroup", "ad", "smart_plus_ad"} {
			rows, err := r.client.list(ctx, kind, "")
			if err != nil {
				addError(&res.Diagnostics, err)
				return
			}
			for _, child := range rows {
				if child["campaign_id"] == id && !deleted(child) {
					addError(&res.Diagnostics, fmt.Errorf("campaign has nondeleted children; deletion blocked"))
					return
				}
			}
		}
		_, err = r.client.call(ctx, r.kind+"_status_update", map[string]any{"advertiser_id": r.client.account, "campaign_ids": []string{id}, "operation_status": "DELETE"})
		if err != nil {
			addError(&res.Diagnostics, err)
			return
		}
		_, err = r.client.verify(ctx, r.kind, id, nil, true)
		if err != nil {
			addError(&res.Diagnostics, err)
			return
		}
	}
	res.State.RemoveResource(ctx)
}
