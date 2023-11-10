variable "task_definition" {
  description = "The task definition resource that is created by the byo-ecs module"
}

variable "ecs_service" {
  description = "The ecs service resource that is created by the byo-ecs module"
}

variable "ecs_cluster" {
  description = "The ecs cluster module that is created by the byo-db module"
}

variable "fleet_config" {
  description = "The root Fleet config object"
  type        = any
}

variable "awslogs_config" {
  type = object({
    group  = string
    region = string
    prefix = string
  })
}


variable "schedule_expression" {
  description = "The scheduled expression in which the cloudwatch target should be executed"
  default     = "rate(1 hour)" // cron(0 * * * *)
}

variable "execution_iam_role_arn" {
  description = "The ARN of the fleet execution role, this is necessary to pass role from ecs events"
}

variable "task_role_arn" {
  description = "The ARN of the fleet task role, this is necessary to pass role from ecs events"
}

variable "vuln_processing_memory" {
  // note must conform to FARGATE breakpoints https://docs.aws.amazon.com/AmazonECS/latest/userguide/fargate-task-defs.html
  default     = 4096
  description = "The amount of memory to dedicate to the vuln processing command"
}

variable "vuln_processing_cpu" {
  // note must conform to FARGETE breakpoints https://docs.aws.amazon.com/AmazonECS/latest/userguide/fargate-task-defs.html
  default     = 1024
  description = "The amount of CPU to dedicate to the vuln processing command"
}

variable "disable_data_sync" {
  default     = "false"
  description = "disable data sync instructs fleet vuln processing to no attempt to pull down new vuln db and artifacts(its expecting everything to already be provisioned at the vulnerabilities database path)"
}