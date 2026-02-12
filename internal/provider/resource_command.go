package provider

import (
	"context"
	"fmt"

	"github.com/ktoulliou/terraform-provider-eon/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &commandResource{}

type commandResource struct{ client *client.Client }

type commandModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	CommandLine types.String `tfsdk:"command_line"`
	Description types.String `tfsdk:"description"`
}

func NewCommandResource() resource.Resource { return &commandResource{} }

func (r *commandResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_command"
}

func (r *commandResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Nagios check command in EON (addCommand / modifyCommand / deleteCommand).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Command name (e.g. check_http).",
			},
			"command_line": schema.StringAttribute{
				Required:    true,
				Description: "Full command line (e.g. $USER1$/check_http -H $HOSTADDRESS$ -p $ARG1$).",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:     stringdefault.StaticString(""),
				Description: "Human-readable description.",
			},
		},
	}
}

func (r *commandResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.Client)
	}
}

func (r *commandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan commandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating EON command", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.AddCommand(
		plan.Name.ValueString(),
		plan.CommandLine.ValueString(),
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating command", err.Error())
		return
	}

	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *commandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state commandModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetCommand(state.Name.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *commandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state commandModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating EON command", map[string]interface{}{"name": state.Name.ValueString()})

	newName := ""
	if plan.Name.ValueString() != state.Name.ValueString() {
		newName = plan.Name.ValueString()
	}

	_, err := r.client.ModifyCommand(
		state.Name.ValueString(),
		newName,
		plan.CommandLine.ValueString(),
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating command", err.Error())
		return
	}

	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *commandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state commandModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting EON command", map[string]interface{}{"name": state.Name.ValueString()})

	_, err := r.client.DeleteCommand(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting command",
			fmt.Sprintf("Could not delete command %q: %s", state.Name.ValueString(), err))
	}
}
