// Copyright github.com/fd008 - All rights reserved
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCloudFrontInvalidation_full(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" {
		t.Skip("AWS_REGION must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			/* add AWS creds check if needed */
		},
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				provider "awsutils" {
					region = "` + os.Getenv("AWS_REGION") + `"
				}

				resource "awsutils_cloudfront_invalidation" "test" {
					distribution_id = "DISTRIBUTION_ID"
					paths           = ["/*"]
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("awsutils_cloudfront_invalidation.test", "invalidation_id"),
				),
			},
		},
	})
}
