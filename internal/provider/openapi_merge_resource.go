// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"terraform-provider-awsutils/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OpenAPIMerge{}
var _ resource.ResourceWithImportState = &OpenAPIMerge{}

func NewOpenAPIMergeResource() resource.Resource {
	return &OpenAPIMerge{}
}

// OpenAPIMerge defines the resource implementation.
type OpenAPIMerge struct{}

// OpenAPIMergeModel describes the resource data model.
type OpenAPIMergeModel struct {
	InputPath  types.String `tfsdk:"input_path"`
	OutputPath types.String `tfsdk:"output_path"`
}

func (r *OpenAPIMerge) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + "merge_openapi_yaml"
}

func (r *OpenAPIMerge) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Merge multiple OpenAPI YAML files into a single file with $ref resolved within local files.",
		Attributes: map[string]schema.Attribute{
			"input_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path to the input OpenAPI files (YAML only).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"output_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path to the output OpenAPI file (YAML only).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}

}

func (r *OpenAPIMerge) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// no-op, no configuration needed
}

func (r *OpenAPIMerge) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// state
	var state OpenAPIMergeModel // Get the planned state from Terraform

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.InputPath.ValueString()
	output := state.OutputPath.ValueString()

	if input == "" || output == "" {
		resp.Diagnostics.AddError(
			"Invalid OpenAPI Merge Configuration",
			fmt.Sprintf("Both input_path %s and output_path %s must be specified.", input, output),
		)
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Merging OpenAPI files from %s to %s", input, output))

	err := utils.OapiYaml(input, output)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error merging OpenAPI files",
			fmt.Sprintf("Could not merge OpenAPI files from %s to %s: %s", state.InputPath.ValueString(), state.OutputPath.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *OpenAPIMerge) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// no-op
}

func (r *OpenAPIMerge) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// no-op, no update logic needed
}

func (r *OpenAPIMerge) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//no-op
}

func (r *OpenAPIMerge) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
