package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type inventoryDataSource struct{ client *apiClient }

func (*inventoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, res *datasource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_inventory"
}
func (*inventoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{Description: "Read-only inventory of regular and Upgraded Smart+ campaigns, ad groups, and ads.", Attributes: map[string]schema.Attribute{
		"advertiser_id":  schema.StringAttribute{Computed: true},
		"resources_json": schema.StringAttribute{Computed: true},
	}}
}
func (d *inventoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, res *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.client, ok = req.ProviderData.(*apiClient)
	if !ok {
		addError(&res.Diagnostics, fmt.Errorf("unexpected provider data"))
	}
}
func (d *inventoryDataSource) Read(ctx context.Context, _ datasource.ReadRequest, res *datasource.ReadResponse) {
	resources := map[string]any{}
	for _, kind := range []string{"campaign", "smart_plus_campaign", "adgroup", "smart_plus_adgroup", "ad", "smart_plus_ad"} {
		rows, err := d.client.list(ctx, kind, "")
		if err != nil {
			addError(&res.Diagnostics, err)
			return
		}
		resources[kind] = rows
	}
	encoded, err := json.Marshal(resources)
	if err != nil {
		addError(&res.Diagnostics, err)
		return
	}
	res.Diagnostics.Append(res.State.Set(ctx, &struct {
		Account   types.String `tfsdk:"advertiser_id"`
		Resources types.String `tfsdk:"resources_json"`
	}{types.StringValue(d.client.account), types.StringValue(string(encoded))})...)
}
