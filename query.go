package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//go:embed queries.json
var querySpecification []byte

type queryOperation struct {
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Schema operation `json:"inputSchema"`
}

var querySpecs = func() map[string]queryOperation {
	var snapshot struct {
		Operations map[string]queryOperation `json:"operations"`
	}
	if err := json.Unmarshal(querySpecification, &snapshot); err != nil {
		panic(err)
	}
	return snapshot.Operations
}()

type queryDataSource struct{ client *apiClient }
type queryModel struct {
	Operation  types.String `tfsdk:"operation"`
	Parameters types.String `tfsdk:"parameters_json"`
	Result     types.String `tfsdk:"result_json"`
}

func (*queryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, res *datasource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_query"
}
func (*queryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{Description: "One read-only request to a pinned, verified GET endpoint. Pagination is explicit; result_json is the single response data object, not an automatically complete collection.", Attributes: map[string]schema.Attribute{
		"operation":       schema.StringAttribute{Required: true, Description: "Operation from docs/coverage.md's read-only query allowlist."},
		"parameters_json": schema.StringAttribute{Optional: true, Description: "JSON request parameters validated against the pinned input schema. Defaults to {}. The configured advertiser is injected where supported."},
		"result_json":     schema.StringAttribute{Computed: true, Sensitive: true, Description: "Raw response data including any pagination metadata. Stored in Terraform state; may include personal or account data."},
	}}
}
func (d *queryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, res *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.client, ok = req.ProviderData.(*apiClient)
	if !ok {
		addError(&res.Diagnostics, fmt.Errorf("unexpected provider data"))
	}
}
func queryParameters(name, encoded, account string) (queryOperation, map[string]any, error) {
	op, ok := querySpecs[name]
	if !ok || op.Method != "GET" {
		return op, nil, fmt.Errorf("unsupported read-only operation %q; see docs/coverage.md", name)
	}
	if encoded == "" {
		encoded = "{}"
	}
	var params map[string]any
	if err := json.Unmarshal([]byte(encoded), &params); err != nil || params == nil {
		return op, nil, fmt.Errorf("parameters_json must be a JSON object")
	}
	for _, key := range []string{"advertiser_id", "advertiser_ids"} {
		if _, exists := op.Schema.Properties[key]; !exists {
			continue
		}
		if v, exists := params[key]; exists {
			if key == "advertiser_id" {
				if v != account {
					return op, nil, fmt.Errorf("advertiser must match provider")
				}
			} else {
				ids, ok := v.([]any)
				if !ok || len(ids) != 1 || ids[0] != account {
					return op, nil, fmt.Errorf("advertiser_ids must contain only the configured advertiser")
				}
			}
		} else if key == "advertiser_id" {
			params[key] = account
		} else {
			for _, required := range op.Schema.Required {
				if required == key {
					params[key] = []any{account}
				}
			}
		}
	}
	if err := validateValue(property{Type: "object", Properties: op.Schema.Properties, Required: op.Schema.Required}, params); err != nil {
		return op, nil, fmt.Errorf("invalid query parameters: %w", err)
	}
	return op, params, nil
}
func (d *queryDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var config queryModel
	res.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if res.Diagnostics.HasError() {
		return
	}
	op, params, err := queryParameters(config.Operation.ValueString(), config.Parameters.ValueString(), d.client.account)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	result, err := d.client.request(ctx, op.Method, op.Path, params)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	config.Result = types.StringValue(string(encoded))
	res.Diagnostics.Append(res.State.Set(ctx, &config)...)
}
