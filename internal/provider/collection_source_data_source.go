package provider

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rockset/rockset-go-client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CollectionSourceDataSource{}

func NewCollectionSourceDataSource() datasource.DataSource {
	return &CollectionSourceDataSource{}
}

// CollectionSourceDataSource defines the data source implementation.
type CollectionSourceDataSource struct {
	client *rockset.RockClient
}

// CollectionSourceDataSourceModel describes the data source data model.
type CollectionSourceDataSourceModel struct {
	Workspace       types.String `tfsdk:"workspace"`
	Collection      types.String `tfsdk:"collection"`
	Id              types.String `tfsdk:"id"`
	IntegrationName types.String `tfsdk:"integration_name"`
	ResumeAt        types.String `tfsdk:"resume_at"`
	SuspendedAt     types.String `tfsdk:"suspended_at"`

	Status *CollectionSourceStatusModel `tfsdk:"status"`
}

type CollectionSourceStatusModel struct {
	State               types.String `tfsdk:"state"`
	Message             types.String `tfsdk:"message"`
	DetectedSizeBytes   types.Number `tfsdk:"detected_size_bytes"`
	LastProcessedAt     types.String `tfsdk:"last_processed_at"`
	LastProcessedItem   types.String `tfsdk:"last_processed_item"`
	TotalProcessedItems types.Number `tfsdk:"total_processed_items"`
}

func (d *CollectionSourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collection_source"
}

func (d *CollectionSourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Rockset collection source data source",
		Attributes: map[string]schema.Attribute{
			"workspace": schema.StringAttribute{
				MarkdownDescription: "Workspace name",
				Required:            true,
			},
			"collection": schema.StringAttribute{
				MarkdownDescription: "Collection name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Source identifier",
				Required:            true,
			},
			"integration_name": schema.StringAttribute{
				MarkdownDescription: "Name of integration",
				Computed:            true,
			},
			"resume_at": schema.StringAttribute{
				MarkdownDescription: "ISO-8601 date when source would be auto resumed, if suspended",
				Computed:            true,
			},
			"suspended_at": schema.StringAttribute{
				MarkdownDescription: "ISO-8601 date when source was suspended, if suspended",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"status": schema.SingleNestedBlock{
				Description: "The ingest status of this source",
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						MarkdownDescription: "Source state",
						Computed:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: "State message",
						Computed:            true,
					},
					"detected_size_bytes": schema.NumberAttribute{
						MarkdownDescription: "Size in bytes detected for the source at collection initialization. This size can be 0 or null for event stream sources.",
						Computed:            true,
					},
					"last_processed_at": schema.StringAttribute{
						MarkdownDescription: "ISO-8601 date when source was last processed",
						Computed:            true,
					},
					"last_processed_item": schema.StringAttribute{
						MarkdownDescription: "Last source item processed by ingester",
						Computed:            true,
					},
					"total_processed_items": schema.NumberAttribute{
						MarkdownDescription: "Total items processed of source",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *CollectionSourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*rockset.RockClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *CollectionSourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CollectionSourceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	workspace := data.Workspace.ValueString()
	collection := data.Collection.ValueString()
	id := data.Id.ValueString()

	request := d.client.SourcesApi.GetSource(ctx, workspace, collection, id)
	response, _, err := request.Execute()
	if err != nil {
		resp.Diagnostics.AddError("GetSource error",
			fmt.Sprintf("Unable to get collection source, got error: %s", err))
		return
	}

	dt := response.GetData()
	data.Workspace = types.StringValue(workspace)
	data.Collection = types.StringValue(collection)
	data.Id = types.StringValue(dt.GetId())
	data.IntegrationName = types.StringValue(dt.GetIntegrationName())
	data.ResumeAt = types.StringValue(dt.GetResumeAt())
	data.SuspendedAt = types.StringValue(dt.GetSuspendedAt())

	if dt.Status != nil {

		data.Status = &CollectionSourceStatusModel{
			State:             types.StringValue(dt.Status.GetState()),
			Message:           types.StringValue(dt.Status.GetMessage()),
			LastProcessedAt:   types.StringValue(dt.Status.GetLastProcessedAt()),
			LastProcessedItem: types.StringValue(dt.Status.GetLastProcessedItem()),
		}

		f := &big.Float{}
		if dt.Status.HasDetectedSizeBytes() {
			f.SetInt64(dt.Status.GetDetectedSizeBytes())
			data.Status.DetectedSizeBytes = types.NumberValue(f)
		}
		if dt.Status.HasTotalProcessedItems() {
			f.SetInt64(dt.Status.GetTotalProcessedItems())
			data.Status.TotalProcessedItems = types.NumberValue(f)
		}
	}

	tflog.Trace(ctx, "read rockset collection source data source",
		map[string]interface{}{
			"workspace":  workspace,
			"collection": collection,
			"id":         id,
		},
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
