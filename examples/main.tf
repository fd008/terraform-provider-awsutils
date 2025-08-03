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


output "config" {
  value = provider::awsutils::sub_data(file("./sub_data.json"), false)
}

output "showlist" {
  value = provider::awsutils::shallow_list("../docs/")
}

output "invalidation" {
  value = awsutils_cloudfront_invalidation.this.invalidation_id
}

