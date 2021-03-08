package terraform.aws.aws_autoscaling_group

violations[rule] {
    input.type = "aws_autoscaling_group"
    input.values.health_check_grace_period > 300

    rule := {
        "ruleID": "TF-AWS-AUTOSCALING-GROUP-0002",
        "ruleDef": "Long health check grace period",
        "level": 2,
        "message": "The health check grace period is too long",
    }
}