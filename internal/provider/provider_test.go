// Copyright github.com/fd008 - All rights reserved
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"awsutils": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the scaffolding provider.
// It allows for testing assertions on data returned by an ephemeral resource during Open.
// The echoprovider is used to arrange tests by echoing ephemeral data into the Terraform state.
// This lets the data be referenced in test assertions with state checks.
// var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
// 	"scaffolding": providerserver.NewProtocol6WithError(New("test")()),
// 	"echo":        echoprovider.NewProviderServer(),
// }

// func testAccPreCheck(t *testing.T) {
// 	// You can add code here to run prior to any test case execution, for example assertions
// 	// about the appropriate environment variables being set are common to see in a pre-check
// 	// function.
// 	if v := testing.Verbose(); v {
// 		t.Log("Running acceptance tests in verbose mode")
// 	}
// 	if v := testing.Short(); v {
// 		t.Log("Running acceptance tests in short mode")
// 	}
// }
