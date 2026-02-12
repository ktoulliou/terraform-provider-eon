package provider

import (
	"context"
	"encoding/json"

	"github.com/eyesofnetwork/terraform-provider-eon/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ═══════════════════════════════════════════════════════════════════════════
//  data "eon_host"
// ═══════════════════════════════════════════════════════════════════════════

var _ datasource.DataSource = &hostDS{}

type hostDS struct{ client *client.Client }

type hostDSModel struct {
	Name       types.String `tfsdk:"name"`
	ResultJSON types.String `tfsdk:"result_json"`
}

func NewHostDataSource() datasource.DataSource { return &hostDS{} }

func (d *hostDS) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *hostDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Nagios host from EON.",
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{Required: true, Description: "Host name to look up."},
			"result_json": schema.StringAttribute{Computed: true, Description: "Raw JSON result from EONAPI."},
		},
	}
}

func (d *hostDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*client.Client)
	}
}

func (d *hostDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg hostDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiResp, err := d.client.GetHost(cfg.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading host", err.Error())
		return
	}
	b, _ := json.Marshal(apiResp.Result)
	cfg.ResultJSON = types.StringValue(string(b))
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

// ═══════════════════════════════════════════════════════════════════════════
//  data "eon_command"
// ═══════════════════════════════════════════════════════════════════════════

var _ datasource.DataSource = &commandDS{}

type commandDS struct{ client *client.Client }

type commandDSModel struct {
	Name       types.String `tfsdk:"name"`
	ResultJSON types.String `tfsdk:"result_json"`
}

func NewCommandDataSource() datasource.DataSource { return &commandDS{} }

func (d *commandDS) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_command"
}

func (d *commandDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Nagios command from EON.",
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{Required: true, Description: "Command name to look up."},
			"result_json": schema.StringAttribute{Computed: true, Description: "Raw JSON result from EONAPI."},
		},
	}
}

func (d *commandDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*client.Client)
	}
}

func (d *commandDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg commandDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiResp, err := d.client.GetCommand(cfg.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading command", err.Error())
		return
	}
	b, _ := json.Marshal(apiResp.Result)
	cfg.ResultJSON = types.StringValue(string(b))
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
