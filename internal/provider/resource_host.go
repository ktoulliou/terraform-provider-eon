package provider

import (
	"context"
	"fmt"

	"github.com/eyesofnetwork/terraform-provider-eon/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &hostResource{}
	_ resource.ResourceWithImportState = &hostResource{}
)

type hostResource struct{ client *client.Client }

type hostModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	IP           types.String `tfsdk:"ip"`
	Alias        types.String `tfsdk:"alias"`
	Template     types.String `tfsdk:"template"`
	Contact      types.String `tfsdk:"contact"`
	ContactGroup types.String `tfsdk:"contact_group"`
	Export       types.Bool   `tfsdk:"export_configuration"`
}

func NewHostResource() resource.Resource { return &hostResource{} }

func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Nagios host in EON (createHost / deleteHost).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Nagios host name (unique identifier).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ip": schema.StringAttribute{
				Required:    true,
				Description: "Host IP address or FQDN.",
			},
			"alias": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:     stringdefault.StaticString(""),
				Description: "Host alias / description.",
			},
			"template": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:     stringdefault.StaticString("GENERIC_HOST"),
				Description: "Parent host template (default: GENERIC_HOST).",
			},
			"contact": schema.StringAttribute{
				Optional:    true,
				Description: "Nagios contact to attach.",
			},
			"contact_group": schema.StringAttribute{
				Optional:    true,
				Description: "Nagios contact group to attach.",
			},
			"export_configuration": schema.BoolAttribute{
				Optional: true, Computed: true,
				Default:     booldefault.StaticBool(false),
				Description: "Reload Nagios config after change (default false). Use eon_export_configuration instead.",
			},
		},
	}
}

func (r *hostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.Client)
	}
}

func (r *hostResource) hostBody(m *hostModel) map[string]interface{} {
	body := map[string]interface{}{
		"templateHostName":    m.Template.ValueString(),
		"hostName":            m.Name.ValueString(),
		"hostIp":              m.IP.ValueString(),
		"hostAlias":           m.Alias.ValueString(),
		"exportConfiguration": m.Export.ValueBool(),
	}
	if !m.Contact.IsNull() && !m.Contact.IsUnknown() {
		body["contactName"] = m.Contact.ValueString()
	}
	if !m.ContactGroup.IsNull() && !m.ContactGroup.IsUnknown() {
		body["contactGroupName"] = m.ContactGroup.ValueString()
	}
	return body
}

func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating EON host", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.CreateHost(r.hostBody(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating host", err.Error())
		return
	}

	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetHost(state.Name.ValueString())
	if err != nil {
		// If 404 / not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}
	// EONAPI getHost doesn't return structured fields we can map back cleanly,
	// so we trust the Terraform state for attribute values.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hostModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating EON host (delete+create)", map[string]interface{}{"name": plan.Name.ValueString()})

	// EONAPI has no modifyHost – delete then recreate
	if _, err := r.client.DeleteHost(plan.Name.ValueString(), false); err != nil {
		resp.Diagnostics.AddError("Error deleting host for update", err.Error())
		return
	}
	if _, err := r.client.CreateHost(r.hostBody(&plan)); err != nil {
		resp.Diagnostics.AddError("Error recreating host", err.Error())
		return
	}

	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting EON host", map[string]interface{}{"name": state.Name.ValueString()})

	_, err := r.client.DeleteHost(state.Name.ValueString(), state.Export.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting host",
			fmt.Sprintf("Could not delete host %q: %s", state.Name.ValueString(), err))
	}
}

func (r *hostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
