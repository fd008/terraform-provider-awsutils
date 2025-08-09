// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	awscloud "terraform-provider-awsutils/internal/aws_cloud"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &kmsPolicyDataSrouce{}
)

// NewKmsPolicyDataSrouce is a helper function to simplify the provider implementation.
func NewKmsPolicyDataSrouce() datasource.DataSource {
	return &kmsPolicyDataSrouce{}
}

// kmsPolicyDataSrouce is the data source implementation.
type kmsPolicyDataSrouce struct {
	KeyID  types.String `tfsdk:"key_id"`
	Policy types.String `tfsdk:"policy"`
	Region types.String `tfsdk:"region"`
	Strict types.Bool   `tfsdk:"strict"`
}

// Metadata returns the data source type name.
func (d *kmsPolicyDataSrouce) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms_policy"
}

// Schema defines the schema for the data source.
func (d *kmsPolicyDataSrouce) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Description of the data source
		Description: "Retrieves the default key policy for a given AWS KMS key.",
		Attributes: map[string]schema.Attribute{
			"key_id": schema.StringAttribute{
				Required:    true, // The key ID is a required input
				Description: "The ID or ARN of the KMS key.",
			},
			"region": schema.StringAttribute{
				Optional:    true, // The region is optional
				Description: "The AWS region where the KMS key is located. If not specified, defaults to the region configured in the provider.",
			},
			"policy": schema.StringAttribute{
				Computed:    true, // The policy is retrieved and thus computed
				Description: "The KMS key policy in JSON format.",
			},
			"strict": schema.BoolAttribute{
				Optional:    true, // The strict mode is optional
				Description: "If true, will throw an error if the key policy is not found. Defaults to false.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *kmsPolicyDataSrouce) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state kmsPolicyDataSrouce

	// Get configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If 'strict' is not set by the user, provide the default value.
	if state.Strict.IsNull() || state.Strict.IsUnknown() {
		state.Strict = types.BoolValue(false) // Set the default value to false
	}

	// Get the KMS Key ID from the configuration
	kmsKeyID := state.KeyID.ValueString()
	region := state.Region.ValueString()

	policy, pErr := awscloud.GetKMSKeyPolicy(kmsKeyID, &region)

	// If the policy retrieval fails and strict mode is enabled, return an error.
	if pErr != nil && state.Strict.ValueBool() {
		resp.Diagnostics.AddError(
			"Error retrieving KMS key policy",
			"Unable to retrieve the KMS key policy for key ID "+kmsKeyID+": "+pErr.Error(),
		)
		return
	}

	// If the policy retrieval fails and strict mode is not enabled, log a warning.
	if pErr != nil && !state.Strict.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Error retrieving KMS key policy",
			"Unable to retrieve the KMS key policy for key ID "+kmsKeyID+": "+pErr.Error(),
		)
		state.Policy = types.StringValue("")
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	state.Policy = types.StringValue(*policy)

	// Set the new state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

}
