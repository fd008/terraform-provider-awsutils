// Copyright (c) https://github.com/fd008 - All rights reserved

terraform {
  required_providers {
    awsutils = {
      source  = "fd008/awsutils"
      version = ">= 1.4.1"
    }
  }
}

provider "awsutils" {
  # region = "us-east-1"
}

locals {
  // Example policies
  policy1JSON = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:aws:iam::123456789012:user/Alice"
        },
        "Action" : [
          "s3:GetObject",
          "s3:ListBucket"
        ],
        "Resource" : ["arn:aws:s3:::my-bucket/*", "arn:aws:s3:::another-bucket/*"]
      },
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Action" : "sqs:SendMessage",
        "Resource" : "arn:aws:sqs:us-east-1:123456789012:my-queue"
      },
      {
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:aws:iam::123456789012:user/Alice"
        },
        "Action" : "s3:GetObject",
        "Resource" : "arn:aws:s3:::my-bucket/object1"
      }
    ]
  })

  policy2JSON = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:aws:iam::123456789012:user/Alice"
        },
        "Action" : "s3:PutObject",
        "Resource" : ["arn:aws:s3:::my-bucket/*"]
      },
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Action" : "sqs:DeleteMessage",
        "Resource" : "arn:aws:sqs:us-east-1:123456789012:my-queue"
      },
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "ec2.amazonaws.com"
        },
        "Action" : "s3:GetBucketLocation",
        "Resource" : "arn:aws:s3:::*"
      },
      {
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:aws:iam::123456789012:user/Alice"
        },
        "Action" : "s3:PutObject",
        "Resource" : [
          "arn:aws:s3:::my-bucket/*"
        ],
        "Condition" : {
          "StringEquals" : {
            "s3:x-amz-acl" : "public-read"
          }
        }
      }
    ]
  })

}

resource "awsutils_cloudfront_invalidation" "this" {
  distribution_id = "ES1234789012"
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



output "iam_policy_merge" {
  value = provider::awsutils::merge_policy(local.policy1JSON, local.policy2JSON)
}
