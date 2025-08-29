// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"terraform-provider-awsutils/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &runCommand{}
var _ resource.ResourceWithImportState = &runCommand{}

func NewRunCommandResource() resource.Resource {
	return &runCommand{}
}

// runCommand defines the resource implementation.
type runCommand struct{}

// runCommandModel describes the resource data model.
type runCommandModel struct {
	ExecFile types.String `tfsdk:"exec_file"`
	Trigger  types.String `tfsdk:"trigger"`
}

func (r *runCommand) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + "run_commands"
}

func (r *runCommand) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Upload a directory to S3 bucket efficiently, with support for large files and concurrency.",
		Attributes: map[string]schema.Attribute{
			"trigger": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Trigger for the upload operation. Setting this to a new value will trigger the upload operation. If not set, defaults to the current timestamp.",
				Description:         "Trigger for the upload operation. Setting this to a new value will trigger the upload operation. If not set, defaults to the current timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"exec_file": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Commands as string to execute, one command per line. Supported commands are: WORKDIR, RUN, COPY, ENV, ENVFILE, and SLEEP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}

}

func (r *runCommand) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// no-op
}

func (r *runCommand) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan runCommandModel // Get the planned state from Terraform

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// randomId, err := generateRandomID(14)
	// if err != nil {
	// 	randomId = time.Now().Format("20060102150405")
	// }

	if plan.Trigger.IsNull() || plan.Trigger.IsUnknown() || plan.Trigger.ValueString() == "" {
		plan.Trigger = types.StringValue(fmt.Sprint(time.Now().Unix()))
	}

	uLang := utils.NewUlang(ctx, plan.ExecFile.ValueString(), "")

	commands, err := uLang.ParseCommands()

	if err != nil {
		resp.Diagnostics.AddError("Error parsing commands", err.Error())
		return
	}

	// Execute each command sequentially
	for _, cmd := range commands {
		tflog.Info(ctx, fmt.Sprintf("--> Executing: %s %s\n", cmd.Name, cmd.Args))

		err := uLang.RunCommand(cmd)
		if err != nil {
			resp.Diagnostics.AddError("Error running command: "+cmd.Name, err.Error())
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *runCommand) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//no-op
}

func (r *runCommand) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// no-op
}

func (r *runCommand) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//no-op
}

func (r *runCommand) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
