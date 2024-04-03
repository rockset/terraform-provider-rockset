package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
	"github.com/rockset/rockset-go-client/wait"
)

// this file should be generated from openapi.Collection

// CollectionResourceModel describes the resource data model.
type CollectionResourceModel struct {
	ID types.String `tfsdk:"id"`
	// require recreation
	Workspace              types.String `tfsdk:"workspace"`
	Name                   types.String `tfsdk:"name"`
	RetentionSecs          types.Int64  `tfsdk:"retention_secs"`
	StorageCompressionType types.String `tfsdk:"storage_compression_type"`
	// works with update
	Description          types.String `tfsdk:"description"`
	IngestTransformation types.String `tfsdk:"ingest_transformation"`
	// optional fields
	WaitForCollection types.Bool `tfsdk:"wait_for_collection"`
	// WaitForDocuments is deprecated and is not used
	WaitForDocuments types.Int64 `tfsdk:"wait_for_documents"`
	// read only fields
	RRN                 types.String `tfsdk:"rrn"`
	ReadOnly            types.Bool   `tfsdk:"read_only"`
	InsertOnly          types.Bool   `tfsdk:"insert_only"`
	CreatedAt           types.String `tfsdk:"created_at"`
	CreatedBy           types.String `tfsdk:"created_by"`
	Status              types.String `tfsdk:"status"`
	CreatedByAPIKeyName types.String `tfsdk:"created_by_apikey_name"`
}

func (m *CollectionResourceModel) Read(response openapi.Collection) {
	m.Workspace = types.StringValue(response.GetWorkspace())
	m.Name = types.StringValue(response.GetName())
	m.Description = types.StringPointerValue(response.Description)
	if response.FieldMappingQuery != nil {
		m.IngestTransformation = types.StringValue(response.FieldMappingQuery.GetSql())
	}
	// TODO(pme) once ORC-4958 is fixed, the nil guard can be removed
	if response.RetentionSecs != nil {
		m.RetentionSecs = types.Int64PointerValue(response.RetentionSecs)
	}
	m.StorageCompressionType = types.StringPointerValue(response.StorageCompressionType)
	m.CreatedAt = types.StringPointerValue(response.CreatedAt)
	m.CreatedBy = types.StringPointerValue(response.CreatedBy)
	m.CreatedByAPIKeyName = types.StringPointerValue(response.CreatedByApikeyName)
	m.InsertOnly = types.BoolPointerValue(response.InsertOnly)
	m.ReadOnly = types.BoolPointerValue(response.ReadOnly)
	m.RRN = types.StringPointerValue(response.Rrn)
	m.Status = types.StringPointerValue(response.Status)
}

func (m *CollectionResourceModel) Options(onlyOnCreate bool) []option.CollectionOption {
	var options []option.CollectionOption

	if !m.Description.IsNull() {
		options = append(options, option.WithCollectionDescription(m.Description.ValueString()))
	}
	if !m.IngestTransformation.IsNull() {
		options = append(options, option.WithIngestTransformation(m.IngestTransformation.ValueString()))
	}

	if onlyOnCreate {
		if !m.RetentionSecs.IsNull() {
			options = append(options, option.WithCollectionRetentionSeconds(m.RetentionSecs.ValueInt64()))
		}
		if !m.StorageCompressionType.IsNull() {
			options = append(options, option.WithStorageCompressionType(option.StorageCompressionType(
				m.StorageCompressionType.ValueString())))
		}
	}

	return options
}

func (m *CollectionResourceModel) Ready(ctx context.Context, rs wait.ResourceGetter) error {
	w := wait.New(rs)
	return w.UntilCollectionReady(ctx, m.Workspace.ValueString(), m.Name.ValueString())
}
