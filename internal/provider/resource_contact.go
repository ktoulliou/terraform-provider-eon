package provider

import (
	"context"
	"fmt"

	"github.com/ktoulliou/terraform-provider-eon/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &contactResource{}

type contactResource struct{ client *client.Client }

type contactModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Alias        types.String `tfsdk:"alias"`
	Mail         types.String `tfsdk:"mail"`
	Pager        types.String `tfsdk:"pager"`
	ContactGroup types.String `tfsdk:"contact_group"`
	Export       types.Bool   `tfsdk:"export_configuration"`
}

func NewContactResource() resource.Resource { return &contactResource{} }

func (r *contactResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func (r *contactResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Nagios contact in EON (createContact / modifyContact / deleteContact).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Contact name.",
			},
			"alias": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:     stringdefault.StaticString(""),
				Description: "Contact alias.",
			},
			"mail": schema.StringAttribute{
				Required:    true,
				Description: "Contact email address.",
			},
			"pager": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:     stringdefault.StaticString(""),
				Description: "Pager number/address.",
			},
			"contact_group": schema.StringAttribute{
				Optional:    true,
				Description: "Contact group to assign this contact to.",
			},
			"export_configuration": schema.BoolAttribute{
				Optional: true, Computed: true,
				Default:     booldefault.StaticBool(false),
				Description: "Reload Nagios config after change.",
			},
		},
	}
}

func (r *contactResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*client.Client)
	}
}

func (r *contactResource) contactBody(m *contactModel) map[string]interface{} {
	body := map[string]interface{}{
		"contactName":         m.Name.ValueString(),
		"contactAlias":        m.Alias.ValueString(),
		"contactMail":         m.Mail.ValueString(),
		"contactPager":        m.Pager.ValueString(),
		"exportConfiguration": m.Export.ValueBool(),
	}
	if !m.ContactGroup.IsNull() && !m.ContactGroup.IsUnknown() {
		body["contactGroup"] = m.ContactGroup.ValueString()
	}
	return body
}

func (r *contactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan contactModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating EON contact", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.CreateContact(r.contactBody(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating contact", err.Error())
		return
	}
	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *contactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state contactModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.GetContact(state.Name.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *contactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state contactModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating EON contact", map[string]interface{}{"name": state.Name.ValueString()})

	body := r.contactBody(&plan)
	body["contactName"] = state.Name.ValueString()
	if plan.Name.ValueString() != state.Name.ValueString() {
		body["newContactName"] = plan.Name.ValueString()
	}

	_, err := r.client.ModifyContact(body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating contact", err.Error())
		return
	}
	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *contactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state contactModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting EON contact", map[string]interface{}{"name": state.Name.ValueString()})

	_, err := r.client.DeleteContact(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting contact",
			fmt.Sprintf("Could not delete contact %q: %s", state.Name.ValueString(), err))
	}
}
