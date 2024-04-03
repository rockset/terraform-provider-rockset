package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rockset/rockset-go-client"
	rockerr "github.com/rockset/rockset-go-client/errors"
)

var (
	_ provider.Provider = &rocksetProvider{}
)

type rocksetProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &rocksetProvider{
			version: version,
		}
	}
}

type RocksetProviderModel struct {
	APIKey    *string `tfsdk:"api_key"`
	APIServer *string `tfsdk:"api_server"`
	OrgID     *string `tfsdk:"organization_id"`
}

const providerUserAgent = "terraform-provider-rockset"

func (p *rocksetProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data RocksetProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var opts = []rockset.RockOption{
		rockset.WithUserAgent(fmt.Sprintf("%s/%s", providerUserAgent, p.version)),
	}

	if data.APIKey != nil {
		opts = append(opts, rockset.WithAPIKey(*data.APIKey))
	}
	if data.APIServer != nil {
		opts = append(opts, rockset.WithAPIServer(*data.APIServer))
	}
	if os.Getenv("ROCKSET_DEBUG") != "" {
		opts = append(opts, rockset.WithHTTPDebug())
	}

	rc, err := rockset.NewClient(opts...)
	if err != nil {
		// TODO create a helper function that turns a Rockset error into a Diagnostic
		resp.Diagnostics.AddError("Failed to create Rockset client", err.Error())
		return
	}

	org, err := rc.GetOrganization(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get organization", err.Error())
		return
	}
	// validate that the expected org id matches the org id of the api key
	if data.OrgID != nil {
		if org.GetId() != *data.OrgID {
			resp.Diagnostics.AddError("Organization ID does not match",
				"The organization configured in the provider does not match the organization of the API key")
			return
		}
	}

	tflog.Info(ctx, "connected to Rockset", map[string]interface{}{"org_id": org.GetId()})

	resp.DataSourceData = rc
	resp.ResourceData = rc
}

func (p *rocksetProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rockset"
	resp.Version = p.version
}

func (p *rocksetProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource {
			return &CollectionSourceDataSource{}
		},
	}
}

func (p *rocksetProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &CollectionResource{} },
	}
}

func (p *rocksetProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The API key used to access Rockset",
				Sensitive:           true,
			},
			"api_server": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The API server for accessing Rockset",
			},
			"organization_id": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The ID of the organization to connect to. " +
					"If this is set, the provider will validate that the organization_id matches the organization_id " +
					"of the api key. If it does not match, the provider will return an error.\n",
			},
		},
	}
}

func rocksetResourceClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *rockset.RockClient {
	client, errDiag := rocksetClient(req.ProviderData)
	if errDiag.Summary() != "" {
		resp.Diagnostics.Append(errDiag)
		return nil
	}

	return client
}

func rocksetDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *rockset.RockClient {
	client, errDiag := rocksetClient(req.ProviderData)
	if errDiag.Summary() != "" {
		resp.Diagnostics.Append(errDiag)
		return nil
	}

	return client
}

func rocksetClient(providerData any) (*rockset.RockClient, diag.ErrorDiagnostic) {
	client, ok := providerData.(*rockset.RockClient)
	if !ok {
		return nil, diag.NewErrorDiagnostic(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *rockset.RockClient, got: %T. Please report this issue to the provider developers.", providerData),
		)
	}

	return client, diag.ErrorDiagnostic{}
}

func DiagFromErr(err error) diag.Diagnostic {
	if err == nil {
		return nil
	}

	var detail string
	var re rockerr.Error
	if errors.As(err, &re) {
		var sb strings.Builder
		var msgs []string

		sb.WriteString(re.GetMessage())
		sb.WriteString(": ")

		if t, ok := re.GetTypeOk(); ok {
			msgs = append(msgs, fmt.Sprintf("Error Type: %s", *t))
		}
		if re.StatusCode != 0 {
			msgs = append(msgs, fmt.Sprintf("HTTP status code (%d) %s", re.StatusCode, http.StatusText(re.StatusCode)))
		}
		if re.GetTraceId() != "" {
			msgs = append(msgs, fmt.Sprintf("Trace ID: %s", re.GetTraceId()))
		}
		if re.GetErrorId() != "" {
			msgs = append(msgs, fmt.Sprintf("Error ID: %s", re.GetErrorId()))
		}
		if re.GetQueryId() != "" {
			msgs = append(msgs, fmt.Sprintf("Query ID: %s", re.GetQueryId()))
		}
		if re.HasLine() {
			msgs = append(msgs, fmt.Sprintf("Line: %d", re.GetLine()))
		}
		if re.HasColumn() {
			msgs = append(msgs, fmt.Sprintf("Column: %d", re.GetColumn()))
		}

		sb.WriteString(strings.Join(msgs, ", "))

		detail = sb.String()
	}

	return diag.NewErrorDiagnostic(err.Error(), detail)
}

func toID(workspace, collection types.String) types.String {
	return types.StringValue(workspace.ValueString() + "." + collection.ValueString())
}
