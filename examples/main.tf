// Copyright (c) https://github.com/fd008 - All rights reserved

terraform {
  required_providers {
    awsutils = {
      source  = "fd008/awsutils"
      version = "1.0.0"
    }
  }
}

provider "awsutils" {
  region = "us-east-1"
}

resource "awsutils_cloudfront_invalidation" "this" {
  distribution_id = "E1EXAMPLE12345"
  paths = [
    "/*"
  ]
  # status = "test"
  trigger = uuid()
}

# resource "awsutils_layer" "node_layer" {
#   dir = "${path.cwd}/fns/one_fn"
#   id  = "one_fn"

# }

# resource "awsutils_layer" "py_layer" {
#   dir = "${path.cwd}/fns/two_fn"
#   id  = "two_fn"

# }

output "config" {
  value = provider::awsutils::sub_data(file("./sub_data.json"), false)
}

# output "invalidation" {
#   value = awsutils_cloudfront_invalidation.this.paths
# }
