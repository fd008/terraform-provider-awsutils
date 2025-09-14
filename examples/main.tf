// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    awsutils = {
      source  = "fd008/awsutils"
      version = ">= 1.7.0"
    }
  }
}

provider "awsutils" {
  # region = "us-east-1"
}

output "filetree" {
  value = provider::awsutils::filetree("../examples", 3, ["*.md", "*.go", "*.tf", ".terraform*"])
}

output "fileset" {
  value = provider::awsutils::fileset("../examples", ["*.yml", "*.git"])
}

output "shallowlist" {
  value = provider::awsutils::shallow_list("../examples", true)

}

# resource "awsutils_s3_dir_upload" "this" {
#   bucket_name = "test-bucket"
#   dir_path    = "../docs"
#   # exclusion_list = [
#   #   "*.md"
#   # ]
#   kms_id  = "some-kms-id-1234"
#   prefix  = "test1/"
#   trigger = uuid()

# }

# resource "awsutils_run_commands" "this" {
#   exec_file = file("${path.module}/execfile")
#   trigger   = uuid()
# }

# data "awsutils_execfile" "this" {
#   exec_file = file("${path.module}/execfile")
#   trigger   = uuid()
# }

# data "awsutils_external" "this" {
#   program     = ["bash", "${path.module}/test.sh"]
#   working_dir = path.module

#   depends_on = [data.awsutils_execfile.this]
# }

# output "files" {
#   value = fileset("./newdir", "**")

#   depends_on = [data.awsutils_external.this]
# }

# resource "awsutils_merge_openapi_yaml" "api" {
#   input_path  = "./api/api.yaml"
#   output_path = "./api/merged.yaml"
# }
# resource "awsutils_cloudfront_invalidation" "this" {
#   distribution_id = "ES1234789012"
#   paths = [
#     "/*"
#   ]
#   # status = "test"
#   trigger = uuid()
# }

# data "awsutils_kms_policy" "this" {
#   key_id = "kms-key-id-12345678901234567890123456789012"
# }

# data "awsutils_s3_policy" "not_exists" {
#   bucket_name = "not-exists-bucket-12345678901234567890123456789012"
#   # region      = "us-east-1"
# }

# data "awsutils_s3_policy" "exists" {
#   bucket_name = "exists-bucket-12345678901234567890123456789012"
#   # region      = "us-east-1"
# }


# output "config" {
#   value = provider::awsutils::sub_data(file("./sub_data.json"), false)
# }


# output "invalidation" {
#   value = awsutils_cloudfront_invalidation.this.invalidation_id
# }

# output "kms" {
#   value = data.awsutils_kms_policy.this.policy
# }

# output "s3_policy_exists" {
#   value = data.awsutils_s3_policy.exists.policy

# }

# output "s3_policy_not_exists" {
#   value = data.awsutils_s3_policy.not_exists.policy

# }



# output "iam_policy_merge" {
#   value = provider::awsutils::merge_policy(local.policy1JSON, local.policy2JSON)
# }
