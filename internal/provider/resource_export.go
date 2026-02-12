package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/ktoulliou/terraform-provider-eon/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &exportConfigResource{}

type exportConfigResource struct{ client *client.Client }

type exportConfigModel struct {
	ID      types.String `tfsdk:"id"`
	JobName types.String `tfsdk:"job_name"`
}

func NewExportConfigResource() resource.Resource { return &exportConfigResource{} }

func (r *exportConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_export_configuration"
}

func (r *exportConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a Nagios configuration export/reload in EON. " +
			"Place this resource last (with depends_on) to apply all changes at once.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"job_name": schema.StringAttribute{
				Required:    true,
				Description: "Export job name (arbitrary label, e.g. 'terraform').",
			},
		},
	}
}

func (r *exportConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.Client)
	}
}

func (r *exportConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan exportConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Exporting Nagios configuration")

	_, err := r.client.ExportConfiguration(plan.JobName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error exporting configuration", err.Error())
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("export-%d", time.Now().Unix()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *exportConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state exportConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *exportConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan exportConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.ExportConfiguration(plan.JobName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error exporting configuration", err.Error())
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("export-%d", time.Now().Unix()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *exportConfigResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Export trigger – nothing to destroy
}
