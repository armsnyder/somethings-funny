provider "aws" {
  region = "${var.region}"
}

data "aws_caller_identity" "default" {}

data "template_file" "container_definitions" {
  template = "${file("container_definitions_tmpl.json")}"
  vars {
    region      = "${var.region}"
    domain      = "${var.domain}"
    bucket_name = "${var.bucket_name}"
    account_id  = "${data.aws_caller_identity.default.account_id}"
    pi_domain   = "${var.pi_domain}"
    hosted_zone = "${var.hosted_zone}"
  }
}

resource "aws_ecs_task_definition" "default" {
  container_definitions    = "${data.template_file.container_definitions.rendered}"
  family                   = "somethings-funny"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = "${var.execution_role_arn}"
  task_role_arn            = "${var.task_role_arn}"
}

data "aws_subnet" "default" {
  id = "${var.subnet_id}"
}

resource "aws_security_group" "default" {
  vpc_id      = "${data.aws_subnet.default.vpc_id}"
  name        = "somethings-funny"
  description = "2019-06-13T08:52:43.153Z"
  ingress {
    protocol    = "tcp"
    from_port   = 443
    to_port     = 443
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_ecs_service" "default" {
  name                               = "somethings-funny"
  desired_count                      = 1
  task_definition                    = "${aws_ecs_task_definition.default.arn}"
  launch_type                        = "FARGATE"
  deployment_maximum_percent         = 100
  deployment_minimum_healthy_percent = 0
  network_configuration {
    subnets          = ["${var.subnet_id}"]
    security_groups  = ["${aws_security_group.default.id}"]
    assign_public_ip = true
  }
}
