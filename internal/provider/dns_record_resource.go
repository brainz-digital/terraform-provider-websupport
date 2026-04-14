package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/brainz-digital/terraform-provider-websupport/internal/client"
)

func NewDNSRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

type dnsRecordResource struct {
	client *client.Client
}

type dnsRecordModel struct {
	ID       types.String `tfsdk:"id"`
	Zone     types.String `tfsdk:"zone"`
	Type     types.String `tfsdk:"type"`
	Name     types.String `tfsdk:"name"`
	Content  types.String `tfsdk:"content"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int64  `tfsdk:"priority"`
	Note     types.String `tfsdk:"note"`
}

func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A DNS record managed via the Websupport REST API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric record ID assigned by Websupport.",
				Computed:    true,
			},
			"zone": schema.StringAttribute{
				Description: "Zone (domain) the record belongs to. Forces replacement.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Record type (A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, etc). Forces replacement.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Record name. Use \"@\" for the apex and \"*\" for wildcard. Forces replacement.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Description: "Record value (IP address, hostname, text, etc).",
				Required:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "TTL in seconds. Defaults to 600.",
				Optional:    true,
				Computed:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "Priority for MX/SRV records.",
				Optional:    true,
			},
			"note": schema.StringAttribute{
				Description: "Free-form note stored alongside the record.",
				Optional:    true,
			},
		},
	}
}

func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func modelToRecord(m *dnsRecordModel) client.Record {
	rec := client.Record{
		Type:    m.Type.ValueString(),
		Name:    m.Name.ValueString(),
		Content: m.Content.ValueString(),
		Note:    m.Note.ValueString(),
	}
	if !m.TTL.IsNull() && !m.TTL.IsUnknown() {
		rec.TTL = int(m.TTL.ValueInt64())
	} else {
		rec.TTL = 600
	}
	if !m.Priority.IsNull() && !m.Priority.IsUnknown() {
		p := int(m.Priority.ValueInt64())
		rec.Priority = &p
	}
	return rec
}

func recordToModel(zone string, rec *client.Record, m *dnsRecordModel) {
	m.ID = types.StringValue(strconv.FormatInt(rec.ID, 10))
	m.Zone = types.StringValue(zone)
	m.Type = types.StringValue(rec.Type)
	m.Name = types.StringValue(rec.Name)
	m.Content = types.StringValue(rec.Content)
	m.TTL = types.Int64Value(int64(rec.TTL))
	if rec.Priority != nil {
		m.Priority = types.Int64Value(int64(*rec.Priority))
	}
	if rec.Note != "" {
		m.Note = types.StringValue(rec.Note)
	}
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := plan.Zone.ValueString()
	created, err := r.client.CreateRecord(zone, modelToRecord(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create DNS record failed", err.Error())
		return
	}

	recordToModel(zone, created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := state.Zone.ValueString()
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid record ID in state", err.Error())
		return
	}

	rec, err := r.client.GetRecord(zone, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read DNS record failed", err.Error())
		return
	}

	recordToModel(zone, rec, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dnsRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid record ID in state", err.Error())
		return
	}
	zone := state.Zone.ValueString()

	updated, err := r.client.UpdateRecord(zone, id, modelToRecord(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update DNS record failed", err.Error())
		return
	}

	recordToModel(zone, updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid record ID in state", err.Error())
		return
	}
	if err := r.client.DeleteRecord(state.Zone.ValueString(), id); err != nil {
		resp.Diagnostics.AddError("Delete DNS record failed", err.Error())
	}
}

// ImportState supports `terraform import websupport_dns_record.foo <zone>/<id>`.
func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: <zone>/<record_id>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
