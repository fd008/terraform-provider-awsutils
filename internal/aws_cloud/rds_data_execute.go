// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ExecuteModel struct {
	ResourceArn           string
	SecretArn             string
	Database              string
	SQL                   string
	Parameters            *map[string]ParameterModel
	Region                *string
	ContinueAfterTimeout  bool
	IncludeResultMetadata bool
}

func (e *ExecuteModel) toInput() (*rdsdata.ExecuteStatementInput, error) {
	var input rdsdata.ExecuteStatementInput

	input.ResourceArn = &e.ResourceArn
	input.SecretArn = &e.SecretArn
	input.Database = &e.Database
	input.Sql = &e.SQL
	input.ContinueAfterTimeout = e.ContinueAfterTimeout || true
	input.IncludeResultMetadata = e.IncludeResultMetadata || true

	tflog.Debug(context.TODO(), fmt.Sprintf("Parameters: %+v", input.Parameters))

	if e.Parameters != nil {
		params, err := toParams(e.Parameters)
		if err != nil {
			return nil, err
		}
		input.Parameters = params
	}

	input.FormatRecordsAs = types.RecordsFormatTypeJson

	return &input, nil
}

func (e *ExecuteModel) ExecuteStatement() (string, error) {
	cfg, err := GetConfig(e.Region, nil, nil)
	if err != nil {
		return "", err
	}

	// Create a new RDS Data Service client
	client := rdsdata.NewFromConfig(cfg)
	input, err := e.toInput()
	if err != nil {
		return "", err
	}

	// Call the ExecuteStatement operation
	resp, err := client.ExecuteStatement(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	// Return the result of the SQL execution
	return anyToJson(resp)
}
