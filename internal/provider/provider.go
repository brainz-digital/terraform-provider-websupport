package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brainz-digital/terraform-provider-websupport/internal/client"
)

type WebsupportProvider struct {
	version string
}

type providerModel struct {
	APIKey    types.String `tfsdk:"api_key"`
	APISecret types.String `tfsdk:"api_secret"`
	BaseURL   types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WebsupportProvider{version: version}
	}
}

func (p *WebsupportProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "websupport"
	resp.Version = p.version
}

func (p *WebsupportProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for Websupport.sk DNS REST API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Websupport API key. Falls back to WEBSUPPORT_API_KEY env var.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_secret": schema.StringAttribute{
				Description: "Websupport API secret. Falls back to WEBSUPPORT_SECRET env var.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "API base URL. Defaults to https://rest.websupport.sk.",
				Optional:    true,
			},
		},
	}
}

func (p *WebsupportProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := data.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("WEBSUPPORT_API_KEY")
	}
	apiSecret := data.APISecret.ValueString()
	if apiSecret == "" {
		apiSecret = os.Getenv("WEBSUPPORT_SECRET")
	}
	baseURL := data.BaseURL.ValueString()

	if apiKey == "" || apiSecret == "" {
		resp.Diagnostics.AddError(
			"Missing Websupport credentials",
			"Set api_key/api_secret in the provider block, or WEBSUPPORT_API_KEY / WEBSUPPORT_SECRET env vars.",
		)
		return
	}

	c := client.New(apiKey, apiSecret, baseURL)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *WebsupportProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSRecordResource,
	}
}

func (p *WebsupportProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
