// Copyright (c) https://github.com/fd008 - All rights reserved

terraform {
  required_providers {
    awsutils = {
      source  = "fd008/awsutils"
      version = ">= 1.5.0"
    }
  }
}

provider "awsutils" {
  # region = "us-east-1"
}


resource "awsutils_s3_dir_upload" "this" {
  bucket_name = "mdt-interop-storage"
  dir_path    = "../docs"
  # exclusion_list = [
  #   "*.md"
  # ]
  kms_id = "9b2fff53-ebdb-4f2e-ab65-c9edc8763978"
  prefix  = "test1/"
  trigger = uuid()

}
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

# output "showlist" {
#   value = provider::awsutils::fileset("..", ["*.yml", "*.git"])
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
