---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: 'bug'
assignees: 'lukaspustina'
---

## Bug Description
A clear and concise description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:
1. Configure Terraform with '...'
2. Run `terraform apply`
3. See error

## Expected Behavior
A clear and concise description of what you expected to happen.

## Actual Behavior
A clear and concise description of what actually happened.

## Environment
- **Provider Version**: [e.g., v0.1.0]
- **Terraform Version**: [e.g., v1.6.0]
- **Pi-hole Version**: [e.g., v5.17.3]
- **Pi-hole API Version**: [e.g., v6]
- **OS**: [e.g., Ubuntu 22.04]

## Configuration Files
```hcl
# Please provide your Terraform configuration
# (sanitize any sensitive information)
terraform {
  required_providers {
    pihole = {
      source = "lukaspustina/pihole"
      version = "~> 0.1"
    }
  }
}

provider "pihole" {
  # Your provider configuration
}

# Your resource configuration that's causing issues
```

## Log Output
```
# Please provide relevant log output
# Run with TF_LOG=DEBUG for more detailed logs
```

## Screenshots
If applicable, add screenshots to help explain your problem.

## Additional Context
Add any other context about the problem here.

## Pi-hole Admin Panel
- [ ] Can you access Pi-hole admin panel successfully?
- [ ] Can you authenticate to the admin panel with the same credentials?
- [ ] Are there any errors in Pi-hole logs?