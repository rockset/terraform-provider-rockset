package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
	"github.com/rockset/rockset-go-client/wait"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.ResourceWithConfigure = &S3SourceResource{}
var _ resource.ResourceWithImportState = &S3SourceResource{}

type S3SourceResource struct {
	client *rockset.RockClient
}

func (r *S3SourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_source"
}

func (r *S3SourceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Rockset collection resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Source identifier",
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
				Description:   "Unique identifier for the collection.",
				Required:      true,
				Validators:    []validator.String{rocksetNameValidator},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"bucket": schema.StringAttribute{
				Description:   "S3 bucket name.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *S3SourceResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.client = rocksetResourceClient(request, response)
}

func (r *S3SourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data S3SourceResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	workspace := data.Workspace.ValueString()
	name := data.Name.ValueString()
	bucket := data.Bucket.ValueString()

	var options []option.CreateS3SourceOption // := data.Options(true)
	_, sourceResponse, err := r.client.CreateS3Source(ctx, workspace, name, bucket, options...)
	if err != nil {
		response.Diagnostics.Append(DiagFromErr(err))
		return
	}

	data.ID = toID(data.Workspace, data.Name)
	data.Read(sourceResponse)

	tflog.Trace(ctx, "created collection", map[string]interface{}{
		"workspace": workspace,
		"name":      name,
	})

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)

	var rm ResourceModel[openapi.SourceS3, option.S3SourceOption] = &data
	if w, ok := rm.(ReadyWaiter); ok {
		if err = w.Ready(ctx, r.client); err != nil {
			response.Diagnostics.Append(DiagFromErr(err))
			return
		}
	}
}

func (r *S3SourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data S3SourceResourceModel
	// Read Terraform prior state data into the model
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	sourceResponse, err := r.client.GetSource(ctx, data.Workspace.ValueString(), data.Name.ValueString(), data.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(DiagFromErr(err))
		return
	}

	data.ID = toID(data.Workspace, data.Name)
	data.Read(*sourceResponse.S3)

	// Save updated data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *S3SourceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data S3SourceResourceModel

	// Read Terraform plan data into the model
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	workspace := data.Workspace.ValueString()
	name := data.Name.ValueString()

	options := data.Options(false)
	_ = options
	collectionResponse, err := r.client.UpdateCollection(ctx, workspace, name)
	if err != nil {
		response.Diagnostics.Append(DiagFromErr(err))
		return
	}

	// data.Read(collectionResponse)
	_ = collectionResponse

	// Save updated data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *S3SourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data S3SourceResourceModel

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

func (r *S3SourceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
