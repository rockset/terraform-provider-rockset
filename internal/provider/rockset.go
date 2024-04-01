package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rockset/rockset-go-client"
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
	return []func() resource.Resource{}
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
