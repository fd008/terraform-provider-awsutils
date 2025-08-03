// Copyright (c) https://github.com/fd008 - All rights reserved

terraform {
  required_providers {
    awsutils = {
      source  = "fd008/awsutils"
      version = "1.1.0"
    }
  }
}

provider "awsutils" {
  # region = "us-east-1"
}

resource "awsutils_cloudfront_invalidation" "this" {
  distribution_id = "ESIMRQH0WC86B"
  paths = [
    "/*"
  ]
  # status = "test"
  trigger = uuid()
}

data "awsutils_kms_policy" "this" {
  key_id = "kms-key-id-12345678901234567890123456789012"
}

data "awsutils_s3_policy" "not_exists" {
  bucket_name = "not-exists-bucket-12345678901234567890123456789012"
  # region      = "us-east-1"
}

data "awsutils_s3_policy" "exists" {
  bucket_name = "exists-bucket-12345678901234567890123456789012"
  # region      = "us-east-1"
}


output "config" {
  value = provider::awsutils::sub_data(file("./sub_data.json"), false)
}

output "showlist" {
  value = provider::awsutils::shallow_list("../docs/")
}

output "invalidation" {
  value = awsutils_cloudfront_invalidation.this.invalidation_id
}

output "kms" {
  value = data.awsutils_kms_policy.this.policy
}

output "s3_policy_exists" {
  value = data.awsutils_s3_policy.exists.policy

}

output "s3_policy_not_exists" {
  value = data.awsutils_s3_policy.not_exists.policy

}
