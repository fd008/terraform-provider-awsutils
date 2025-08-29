// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"terraform-provider-awsutils/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ExecfileDataSource{}

func NewExecfileDataSource() datasource.DataSource {
	return &ExecfileDataSource{}
}

// ExecfileDataSource defines the data source implementation.
type ExecfileDataSource struct {
}

// ExecfileDataSourceModel describes the data source data model.
type ExecfileDataSourceModel struct {
	ExecFile types.String `tfsdk:"exec_file"`
	Trigger  types.String `tfsdk:"trigger"`
}

func (d *ExecfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_execfile"
}

func (d *ExecfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Execfile data source",

		Attributes: map[string]schema.Attribute{
			"trigger": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Trigger for the upload operation. Setting this to a new value will trigger the upload operation. If not set, defaults to the current timestamp.",
				Description:         "Trigger for the upload operation. Setting this to a new value will trigger the upload operation. If not set, defaults to the current timestamp.",
			},
			"exec_file": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Commands as string to execute, one command per line. Supported commands are: WORKDIR, RUN, COPY, ENV, ENVFILE, and SLEEP.",
			},
		},
	}
}

func (d *ExecfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

}

func (d *ExecfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan runCommandModel // Get the planned state from Terraform

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
