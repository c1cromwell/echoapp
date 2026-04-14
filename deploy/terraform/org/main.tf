terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Remote state — uncomment after S3 bucket is created
  # backend "s3" {
  #   bucket  = "echo-terraform-state"
  #   key     = "org/terraform.tfstate"
  #   region  = "us-east-1"
  #   encrypt = true
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "echo"
      ManagedBy   = "terraform"
      Environment = "management"
    }
  }
}

# -------------------------------------------------------------------
# Variables
# -------------------------------------------------------------------

variable "aws_region" {
  description = "Primary AWS region"
  type        = string
  default     = "us-east-1"
}

variable "org_email_prefix" {
  description = "Email prefix for sub-accounts (e.g., 'chad' creates chad+echo-dev@gmail.com)"
  type        = string
}

variable "org_email_domain" {
  description = "Email domain for sub-accounts"
  type        = string
  default     = "gmail.com"
}

# -------------------------------------------------------------------
# AWS Organization
# -------------------------------------------------------------------

resource "aws_organizations_organization" "echo" {
  feature_set = "ALL"

  aws_service_access_principals = [
    "sso.amazonaws.com",
    "cloudtrail.amazonaws.com",
  ]

  enabled_policy_types = [
    "SERVICE_CONTROL_POLICY",
  ]
}

# -------------------------------------------------------------------
# Organizational Units
# -------------------------------------------------------------------

resource "aws_organizations_organizational_unit" "workloads" {
  name      = "Workloads"
  parent_id = aws_organizations_organization.echo.roots[0].id
}

# -------------------------------------------------------------------
# Sub-Accounts: dev, staging, prod
# -------------------------------------------------------------------

resource "aws_organizations_account" "dev" {
  name      = "echo-dev"
  email     = "${var.org_email_prefix}+echo-dev@${var.org_email_domain}"
  parent_id = aws_organizations_organizational_unit.workloads.id
  role_name = "EchoAdminRole"

  tags = { Environment = "development" }

  lifecycle {
    ignore_changes = [role_name]
  }
}

resource "aws_organizations_account" "staging" {
  name      = "echo-staging"
  email     = "${var.org_email_prefix}+echo-staging@${var.org_email_domain}"
  parent_id = aws_organizations_organizational_unit.workloads.id
  role_name = "EchoAdminRole"

  tags = { Environment = "staging" }

  lifecycle {
    ignore_changes = [role_name]
  }
}

resource "aws_organizations_account" "prod" {
  name      = "echo-prod"
  email     = "${var.org_email_prefix}+echo-prod@${var.org_email_domain}"
  parent_id = aws_organizations_organizational_unit.workloads.id
  role_name = "EchoAdminRole"

  tags = { Environment = "production" }

  lifecycle {
    ignore_changes = [role_name]
  }
}

# -------------------------------------------------------------------
# Outputs
# -------------------------------------------------------------------

output "organization_id" {
  value = aws_organizations_organization.echo.id
}

output "dev_account_id" {
  value = aws_organizations_account.dev.id
}

output "staging_account_id" {
  value = aws_organizations_account.staging.id
}

output "prod_account_id" {
  value = aws_organizations_account.prod.id
}
