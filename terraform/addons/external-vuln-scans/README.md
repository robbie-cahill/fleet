# External Vulnerability Scans addon
This addon moves vulnerability scans off of serving nodes and onto a scheduled task in AWS Eventbridge.

Usage is simplified by using the output from the fleet byo-ecs module (../terraform/byo-vpc/byo-db/byo-ecs/README.md)

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | 5.24.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 5.11.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_event_rule.main](https://registry.terraform.io/providers/hashicorp/aws/5.24.0/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_target.ecs_scheduled_task](https://registry.terraform.io/providers/hashicorp/aws/5.24.0/docs/resources/cloudwatch_event_target) | resource |
| [aws_ecs_task_definition.vuln-processing](https://registry.terraform.io/providers/hashicorp/aws/5.24.0/docs/resources/ecs_task_definition) | resource |
| [aws_iam_role.ecs_events](https://registry.terraform.io/providers/hashicorp/aws/5.24.0/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.ecs_events_run_task_with_any_role](https://registry.terraform.io/providers/hashicorp/aws/5.24.0/docs/resources/iam_role_policy) | resource |
| [aws_iam_policy_document.assume_role](https://registry.terraform.io/providers/hashicorp/aws/5.24.0/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.ecs_events_run_task_with_any_role](https://registry.terraform.io/providers/hashicorp/aws/5.24.0/docs/data-sources/iam_policy_document) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/5.24.0/docs/data-sources/region) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_ecs_cluster"></a> [ecs\_cluster](#input\_ecs\_cluster) | The ecs cluster module that is created by the byo-db module | `any` | n/a | yes |
| <a name="input_ecs_service"></a> [ecs\_service](#input\_ecs\_service) | The ecs service resource that is created by the byo-ecs module | `any` | n/a | yes |
| <a name="input_execution_iam_role_arn"></a> [execution\_iam\_role\_arn](#input\_execution\_iam\_role\_arn) | The ARN of the fleet execution role, this is necessary to pass role from ecs events | `any` | n/a | yes |
| <a name="input_fleet_config"></a> [fleet\_config](#input\_fleet\_config) | The root Fleet config object | `any` | n/a | yes |
| <a name="input_logging_options"></a> [logging\_options](#input\_logging\_options) | n/a | `map(string)` | n/a | yes |
| <a name="input_schedule_expression"></a> [schedule\_expression](#input\_schedule\_expression) | The scheduled expression in which the cloudwatch target should be executed | `string` | `"rate(1 hour)"` | no |
| <a name="input_task_definition"></a> [task\_definition](#input\_task\_definition) | The task definition resource that is created by the byo-ecs module | `any` | n/a | yes |
| <a name="input_task_role_arn"></a> [task\_role\_arn](#input\_task\_role\_arn) | The ARN of the fleet task role, this is necessary to pass role from ecs events | `any` | n/a | yes |
| <a name="input_vuln_processing_cpu"></a> [vuln\_processing\_cpu](#input\_vuln\_processing\_cpu) | The amount of CPU to dedicate to the vuln processing command | `number` | `1024` | no |
| <a name="input_vuln_processing_memory"></a> [vuln\_processing\_memory](#input\_vuln\_processing\_memory) | The amount of memory to dedicate to the vuln processing command | `number` | `4096` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_extra_environment_variables"></a> [extra\_environment\_variables](#output\_extra\_environment\_variables) | n/a |
