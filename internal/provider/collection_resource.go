package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
	"github.com/rockset/rockset-go-client/wait"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.ResourceWithConfigure = &CollectionResource{}
var _ resource.ResourceWithImportState = &CollectionResource{}

type ReadyWaiter interface {
	Ready(context.Context, wait.ResourceGetter) error
}

type ResourceModel[R any, O any] interface {
	Read(resource R)
	Options(bool) []O
}

// CollectionResource defines the resource implementation.
type CollectionResource struct {
	client *rockset.RockClient
}

func (r *CollectionResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_collection"
}

func (r *CollectionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Rockset collection resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Collection identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace": schema.StringAttribute{
				Description:   "The name of the workspace.",
				Required:      true,
				Validators:    []validator.String{rocksetNameValidator},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Description:   "Unique identifier for the collection. Can contain alphanumeric or dash characters.",
				Required:      true,
				Validators:    []validator.String{rocksetNameValidator},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Description: "Created by the Rockset terraform provider.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("created by Rockset terraform provider"),
			},
			"ingest_transformation": schema.StringAttribute{
				Description: `Ingest transformation SQL query. Turns the collection into insert_only mode.

When inserting data into Rockset, you can transform the data by providing a single SQL query, 
that contains all of the desired data transformations. 
This is referred to as the collection’s ingest transformation or, historically, its field mapping query.

For more information see https://docs.rockset.com/documentation/docs/ingest-transformation`,
				Optional: true,
				// TODO we need a plan modifier to see if we can detect if the change requires a replace
				//  as only certain changes can be done without a replace
			},
			"retention_secs": schema.Int64Attribute{
				Description:   "Number of seconds after which data is purged based on event time.",
				Optional:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators: []validator.Int64{
					int64validator.Any(
						// according to the docs, 1h is the min value
						int64validator.Between(0, 0),
						int64validator.Between(3_600, 315_360_000),
					),
				},
			},
			"storage_compression_type": schema.StringAttribute{
				Description: "RocksDB storage compression type. Possible values: ZSTD, LZ4.",
				// we can't pick a default value, as the default depends on the org settings
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
				Validators: []validator.String{
					stringvalidator.OneOf(option.StorageCompressionZSTD, option.StorageCompressionLZ4),
				},
			},
			// TODO(pme) add support in the client for this
			// "source_download_soft_limit_bytes": schema.Int64Attribute{
			// 	Description: "Soft ingest limit for this collection.",
			// 	Optional:    true,
			// },
			// options
			"wait_for_collection": schema.BoolAttribute{
				Description: "Wait until the collection is ready.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"wait_for_documents": schema.Int64Attribute{
				Description:        "Wait until the collection has documents. The default is to wait for 0 documents, which means it doesn't wait.",
				Optional:           true,
				DeprecationMessage: "This attribute is deprecated and will be removed in a future release.",
			},
			// read only attributes below
			"created_at": schema.StringAttribute{
				Description: "The time the collection was created.",
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "Email of user who created the collection.",
				Computed:    true,
			},
			"created_by_apikey_name": schema.StringAttribute{
				Description: "Name of the API key that was used to create this collection if one was used.",
				Computed:    true,
			},
			"insert_only": schema.BoolAttribute{
				Description: "Whether the collection is insert only.",
				Computed:    true,
			},
			"read_only": schema.BoolAttribute{
				Description: "Whether the collection is read only.",
				Computed:    true,
			},
			"rrn": schema.StringAttribute{
				Description: "Rockset resource name.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the collection.",
				Computed:    true,
			},
		},
	}
}

func (r *CollectionResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.client = rocksetResourceClient(request, response)
}

func (r *CollectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data CollectionResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	workspace := data.Workspace.ValueString()
	name := data.Name.ValueString()

	options := data.Options(true)
	collectionResponse, err := r.client.CreateCollection(ctx, workspace, name, options...)
	if err != nil {
		response.Diagnostics.Append(DiagFromErr(err))
		return
	}

	data.ID = toID(data.Workspace, data.Name)
	data.Read(collectionResponse)

	tflog.Trace(ctx, "created collection", map[string]interface{}{
		"workspace": workspace,
		"name":      name,
	})

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)

	var rm ResourceModel[openapi.Collection, option.CollectionOption] = &data
	if w, ok := rm.(ReadyWaiter); ok {
		if err = w.Ready(ctx, r.client); err != nil {
			response.Diagnostics.Append(DiagFromErr(err))
			return
		}
	}
}

func (r *CollectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data CollectionResourceModel

	// Read Terraform prior state data into the model
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	collectionResponse, err := r.client.GetCollection(ctx, data.Workspace.ValueString(), data.Name.ValueString())
	if err != nil {
		response.Diagnostics.Append(DiagFromErr(err))
		return
	}

	data.ID = toID(data.Workspace, data.Name)
	data.Read(collectionResponse)

	// Save updated data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data CollectionResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	workspace := data.Workspace.ValueString()
	name := data.Name.ValueString()

	options := data.Options(false)
	collectionResponse, err := r.client.UpdateCollection(ctx, workspace, name, options...)
	if err != nil {
		response.Diagnostics.Append(DiagFromErr(err))
		return
	}

	data.Read(collectionResponse)

	// Save updated data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *CollectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data CollectionResourceModel

	// Read Terraform prior state data into the model
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCollection(ctx, data.Workspace.ValueString(), data.Name.ValueString()); err != nil {
		response.Diagnostics.Append(DiagFromErr(err))
		return
	}

	w := wait.New(r.client)
	if err := w.UntilCollectionGone(ctx, data.Workspace.ValueString(), data.Name.ValueString()); err != nil {
		response.Diagnostics.Append(DiagFromErr(err))
		return
	}
}

func (r *CollectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
