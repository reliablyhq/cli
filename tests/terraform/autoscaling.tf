# Configure the AWS Provider
provider "aws" {
  region = "us-east-1"
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = "ami-1a2b3c"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-east-1a"]
  desired_capacity   = 1
  max_size           = 1
  min_size           = 1
  health_check_grace_period = 600
  value = afunc(cats)

  launch_template {
    id      = aws_launch_template.foobar.id
    version = "$Latest"
  }
}
