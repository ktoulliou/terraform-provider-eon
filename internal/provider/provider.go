package provider

import (
	"context"
	"os"

	"github.com/ktoulliou/terraform-provider-eon/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &eonProvider{}

type eonProvider struct{ version string }

type eonProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	APIKey   types.String `tfsdk:"api_key"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &eonProvider{version: version} }
}

func (p *eonProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "eon"
	resp.Version = p.version
}

func (p *eonProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for EyesOfNetwork – manages Nagios hosts, check commands and contacts via EONAPI.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional:    true,
				Description: "EONAPI base URL (e.g. https://eon.example.com/eonapi). Env: EON_URL.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Admin username. Env: EON_USERNAME.",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "EONAPI key (from /getApiKey). Env: EON_API_KEY.",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS verification (default false).",
			},
		},
	}
}

func envOrVal(val types.String, envKey string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envKey)
}

func (p *eonProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg eonProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eonURL := envOrVal(cfg.URL, "EON_URL")
	username := envOrVal(cfg.Username, "EON_USERNAME")
	apiKey := envOrVal(cfg.APIKey, "EON_API_KEY")

	if eonURL == "" {
		resp.Diagnostics.AddError("Missing url", "Set 'url' in provider block or EON_URL env var.")
		return
	}
	if username == "" {
		resp.Diagnostics.AddError("Missing username", "Set 'username' in provider block or EON_USERNAME env var.")
		return
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("Missing api_key", "Set 'api_key' in provider block or EON_API_KEY env var.")
		return
	}

	insecure := false
	if !cfg.Insecure.IsNull() {
		insecure = cfg.Insecure.ValueBool()
	}

	c := client.NewClient(eonURL, username, apiKey, insecure)
	if err := c.CheckAuth(); err != nil {
		resp.Diagnostics.AddError("EONAPI authentication failed", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *eonProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHostResource,
		NewCommandResource,
		NewContactResource,
		NewContactGroupResource,
		NewExportConfigResource,
	}
}

func (p *eonProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHostDataSource,
		NewCommandDataSource,
	}
}
