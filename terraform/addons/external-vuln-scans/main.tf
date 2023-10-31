data "aws_region" "current" {}

locals {
  environment = [for k, v in var.fleet_config.extra_environment_variables : {
    name  = k
    value = v
  }]
  secrets = [for k, v in var.fleet_config.extra_secrets : {
    name      = k
    valueFrom = v
  }]
}

resource "aws_cloudwatch_event_rule" "main" {
  schedule_expression = var.schedule_expression
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com", "ecs-tasks.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "ecs_events" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "ecs_events_run_task_with_any_role" {
  statement {
    effect    = "Allow"
    actions   = ["iam:PassRole"]
    resources = [var.execution_iam_role_arn]
  }

  statement {
    effect    = "Allow"
    actions   = ["ecs:RunTask"]
    resources = [replace(var.task_definition.arn, "/:\\d+$/", ":*")]
    condition {
      test     = "ArnEquals"
      values   = [var.ecs_cluster.cluster_arn]
      variable = "ecs:cluster"
    }
  }
}
resource "aws_iam_role_policy" "ecs_events_run_task_with_any_role" {
  role   = aws_iam_role.ecs_events.id
  policy = data.aws_iam_policy_document.ecs_events_run_task_with_any_role.json
}

resource "aws_ecs_task_definition" "vuln-processing" {
  family                   = var.fleet_config.family
  cpu                      = var.fleet_config.vuln_processing_cpu
  memory                   = var.fleet_config.vuln_processing_mem
  execution_role_arn       = aws_iam_role.execution.arn
  task_role_arn            = aws_iam_role.main.arn
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]

  container_definitions = jsonencode([
    {
      name        = "fleet-vuln-processing"
      image       = var.fleet_config.image
      essential   = true
      command     = ["fleet", "vuln_processing"]
      user        = "root"
      networkMode = "awsvpc"
      mountPoints = [
        {
          sourceVolume  = "efs-mount"
          containerPath = var.fleet_config.vuln_database_path
          readOnly      = false
        }
      ],
      secrets = concat(
        [
          {
            name      = "FLEET_MYSQL_PASSWORD"
            valueFrom = var.fleet_config.database.password_secret_arn
          }
        ], local.secrets),
      environment = concat(
        [
          {
            name  = "FLEET_MYSQL_USERNAME"
            value = var.fleet_config.database.user
          },
          {
            name  = "FLEET_MYSQL_DATABASE"
            value = var.fleet_config.database.database
          },
          {
            name  = "FLEET_MYSQL_ADDRESS"
            value = var.fleet_config.database.address
          },
          {
            name  = "FLEET_VULNERABILITIES_DISABLE_DATA_SYNC"
            value = "true"
          },
          {
            name  = "FLEET_VULNERABILITIES_DATABASES_PATH"
            value = var.fleet_config.vuln_database_path
          }
        ], local.environment),
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = var.fleet_config.awslogs.name
          awslogs-region        = var.fleet_config.awslogs.region
          awslogs-stream-prefix = "${var.fleet_config.awslogs.prefix}-vuln-procssing"
        }
      }
    }
  ])
}

resource "aws_cloudwatch_event_target" "ecs_scheduled_task" {
  arn      = var.ecs_cluster.cluster_arn
  rule     = aws_cloudwatch_event_rule.main.name
  role_arn = aws_iam_role.ecs_events.arn

  ecs_target {
    task_count          = 1
    task_definition_arn = var.task_definition.arn
    launch_type         = "FARGATE"
    network_configuration {
      subnets         = var.ecs_service.network_configuration[0].subnets
      security_groups = var.ecs_service.network_configuration[0].security_groups
    }
  }

  input = jsonencode({
    containerOverrides = [
      {
        name    = "fleet",
        command = ["fleet", "vuln_processing"]
        cpu     = var.vuln_processing_cpu,
        memory  = var.vuln_processing_memory,
      },
    ]
  })
}
