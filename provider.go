package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tiktokProvider struct{ testClient *apiClient }

func newProvider() provider.Provider { return &tiktokProvider{} }
func (*tiktokProvider) Metadata(_ context.Context, _ provider.MetadataRequest, r *provider.MetadataResponse) {
	r.TypeName, r.Version = "tiktok", "0.3.0"
}
func (*tiktokProvider) Schema(_ context.Context, _ provider.SchemaRequest, r *provider.SchemaResponse) {
	r.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"advertiser_id": schema.StringAttribute{Required: true, Description: "TikTok advertiser ID. Credentials come only from TIKTOK_MARKETING_ACCESS_TOKEN."},
		"allow_writes":  schema.BoolAttribute{Optional: true, Description: "Explicitly enable remote mutations. Defaults to false."},
		"allow_delete":  schema.BoolAttribute{Optional: true, Description: "Explicitly enable terminal campaign deletion. Defaults to false."},
	}}
}
func (p *tiktokProvider) Configure(ctx context.Context, req provider.ConfigureRequest, r *provider.ConfigureResponse) {
	var config struct {
		Account types.String `tfsdk:"advertiser_id"`
		Writes  types.Bool   `tfsdk:"allow_writes"`
		Delete  types.Bool   `tfsdk:"allow_delete"`
	}
	r.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if r.Diagnostics.HasError() {
		return
	}
	if config.Account.IsUnknown() || !validID(config.Account.ValueString()) {
		r.Diagnostics.AddError("Invalid advertiser", "advertiser_id must be a known decimal string.")
		return
	}
	client := &apiClient{account: config.Account.ValueString(), token: os.Getenv("TIKTOK_MARKETING_ACCESS_TOKEN"), writes: config.Writes.ValueBool(), allowDelete: config.Delete.ValueBool(), base: "https://business-api.tiktok.com", http: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return fmt.Errorf("redirect refused") }}}
	if p.testClient != nil {
		client = p.testClient
	}
	r.ResourceData, r.DataSourceData = client, client
}
func (*tiktokProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &campaignResource{kind: "campaign"} },
		func() resource.Resource { return &campaignResource{kind: "smart_plus_campaign"} },
		func() resource.Resource { return &objectResource{kind: "adgroup"} },
		func() resource.Resource { return &objectResource{kind: "ad"} },
		func() resource.Resource { return &objectResource{kind: "smart_plus_adgroup"} },
		func() resource.Resource { return &objectResource{kind: "smart_plus_ad"} },
	}
}
func (*tiktokProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{func() datasource.DataSource { return &inventoryDataSource{} }}
}
