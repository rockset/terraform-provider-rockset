package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
	"github.com/rockset/rockset-go-client/wait"
)

type S3SourceResourceModel struct {
	ID types.String `tfsdk:"id"`
	// require recreation
	Workspace types.String `tfsdk:"workspace"`
	Name      types.String `tfsdk:"name"`
	Bucket    types.String `tfsdk:"bucket"`
	Region    types.String `tfsdk:"region"`
	Pattern   types.String `tfsdk:"pattern"`
	// works with update
	// s3_scan_frequency
	// read only fields
	ObjectBytesDownloaded types.Int64 `tfsdk:"object_bytes_downloaded"`
	ObjectBytesTotal      types.Int64 `tfsdk:"object_bytes_total"`
	ObjectCountDownloaded types.Int64 `tfsdk:"object_count_downloaded"`
	ObjectCountTotal      types.Int64 `tfsdk:"object_count_total"`
}

func (m *S3SourceResourceModel) Read(response openapi.SourceS3) {
	m.Bucket = types.StringValue(response.GetBucket())
}

func (m *S3SourceResourceModel) Ready(ctx context.Context, rs wait.ResourceGetter) error {
	w := wait.New(rs)
	return w.UntilSourceProcessing(ctx, m.Workspace.ValueString(), m.Name.ValueString(), m.ID.ValueString())
}

func (m *S3SourceResourceModel) Options(onlyOnCreate bool) []option.S3SourceOption {
	return nil
}
