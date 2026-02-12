package provider

import (
	"context"
	"fmt"

	"github.com/eyesofnetwork/terraform-provider-eon/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &contactGroupResource{}

type contactGroupResource struct{ client *client.Client }

type contactGroupModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Export      types.Bool   `tfsdk:"export_configuration"`
}

func NewContactGroupResource() resource.Resource { return &contactGroupResource{} }

func (r *contactGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact_group"
}

func (r *contactGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Nagios contact group in EON.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Contact group name.",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:     stringdefault.StaticString("contact group"),
				Description: "Group description.",
			},
			"export_configuration": schema.BoolAttribute{
				Optional: true, Computed: true,
				Default: booldefault.StaticBool(false),
			},
		},
	}
}

func (r *contactGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.Client)
	}
}

func (r *contactGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan contactGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.CreateContactGroup(
		plan.Name.ValueString(), plan.Description.ValueString(), plan.Export.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating contact group", err.Error())
		return
	}
	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *contactGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state contactGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.GetContactGroup(state.Name.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *contactGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state contactGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"contactGroupName":    state.Name.ValueString(),
		"description":         plan.Description.ValueString(),
		"exportConfiguration": plan.Export.ValueBool(),
	}
	if plan.Name.ValueString() != state.Name.ValueString() {
		body["newContactGroupName"] = plan.Name.ValueString()
	}
	_, err := r.client.ModifyContactGroup(body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating contact group", err.Error())
		return
	}
	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *contactGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state contactGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.DeleteContactGroup(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting contact group",
			fmt.Sprintf("Could not delete %q: %s", state.Name.ValueString(), err))
	}
}
