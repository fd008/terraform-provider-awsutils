// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
)

type ExecuteBatchModel struct {
	ResourceArn string
	SecretArn   string
	Database    string
	SQL         string
	Parameters  [][]types.SqlParameter
	Region      *string
}

func (e *ExecuteBatchModel) ExecuteBatchStatement() (string, error) {
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
	resp, err := client.BatchExecuteStatement(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to execute batch SQL statement: %w", err)
	}

	// Return the result of the SQL execution
	return anyToJson(resp)
}
