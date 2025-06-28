// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &AWSUtilsProvider{}
var _ provider.ProviderWithFunctions = &AWSUtilsProvider{}
var _ provider.ProviderWithEphemeralResources = &AWSUtilsProvider{}

type AWSUtilsProvider struct {
	version string
}

type AWSUtilsProviderModel struct {
	Region                 *string  `tfsdk:"region"`
	SharedConfigFiles      []string `tfsdk:"shared_config_files"`
	SharedCredentialsFiles []string `tfsdk:"shared_credentials_files"`
}

func (p *AWSUtilsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "awsutils"
	resp.Version = p.version
}

func (p *AWSUtilsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "The region to use for AWS requests.",
				Optional:    true,
			},
			"shared_config_files": schema.ListAttribute{
				Description: "Path to shared config file. If not set, defaults to ~/.aws/config.",
				Optional:    true,
				ElementType: types.StringNull().Type(context.TODO()),
			},
			"shared_credentials_files": schema.ListAttribute{
				Description: "List of paths to shared credentials files. If not set, defaults to [~/.aws/credentials].",
				Optional:    true,
				ElementType: types.StringNull().Type(context.TODO()),
			},
		},
	}
}

func (p *AWSUtilsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AWSUtilsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AWSUtilsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCfResource,
	}
}

func (p *AWSUtilsProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *AWSUtilsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *AWSUtilsProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewAwsVarFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AWSUtilsProvider{
			version: version,
		}
	}
}
