package scan

import (
	"encoding/json"
	"testing"

	"github.com/reliablyhq/cli/core/terraform"
	"github.com/stretchr/testify/assert"
)

var (
	// target
	testTarget = &Target{
		ResourceType: "aws_autoscaling_group",
		Platform:     "terraform",
	}

	terraformTestPlan = []byte(`{"format_version":"0.1","terraform_version":"0.14.7","planned_values":{"root_module":{"resources":[{"address":"aws_autoscaling_group.bar","mode":"managed","type":"aws_autoscaling_group","name":"bar","provider_name":"registry.terraform.io/hashicorp/aws","schema_version":0,"values":{"availability_zones":["us-east-1a"],"capacity_rebalance":null,"desired_capacity":1,"enabled_metrics":null,"force_delete":false,"health_check_grace_period":600,"initial_lifecycle_hook":[],"instance_refresh":[],"launch_configuration":null,"launch_template":[{"version":"$Latest"}],"load_balancers":null,"max_instance_lifetime":null,"max_size":1,"metrics_granularity":"1Minute","min_elb_capacity":null,"min_size":1,"mixed_instances_policy":[],"name_prefix":null,"placement_group":null,"protect_from_scale_in":false,"suspended_processes":null,"tag":[],"tags":null,"target_group_arns":null,"termination_policies":null,"timeouts":null,"wait_for_capacity_timeout":"10m","wait_for_elb_capacity":null}},{"address":"aws_launch_template.foobar","mode":"managed","type":"aws_launch_template","name":"foobar","provider_name":"registry.terraform.io/hashicorp/aws","schema_version":0,"values":{"block_device_mappings":[],"capacity_reservation_specification":[],"cpu_options":[],"credit_specification":[],"description":null,"disable_api_termination":null,"ebs_optimized":null,"elastic_gpu_specifications":[],"elastic_inference_accelerator":[],"enclave_options":[],"hibernation_options":[],"iam_instance_profile":[],"image_id":"ami-1a2b3c","instance_initiated_shutdown_behavior":null,"instance_market_options":[],"instance_type":"t2.micro","kernel_id":null,"key_name":null,"license_specification":[],"monitoring":[],"name_prefix":"foobar","network_interfaces":[],"placement":[],"ram_disk_id":null,"security_group_names":null,"tag_specifications":[],"tags":null,"update_default_version":null,"user_data":null,"vpc_security_group_ids":null}}]}},"resource_changes":[{"address":"aws_autoscaling_group.bar","mode":"managed","type":"aws_autoscaling_group","name":"bar","provider_name":"registry.terraform.io/hashicorp/aws","change":{"actions":["create"],"before":null,"after":{"availability_zones":["us-east-1a"],"capacity_rebalance":null,"desired_capacity":1,"enabled_metrics":null,"force_delete":false,"health_check_grace_period":600,"initial_lifecycle_hook":[],"instance_refresh":[],"launch_configuration":null,"launch_template":[{"version":"$Latest"}],"load_balancers":null,"max_instance_lifetime":null,"max_size":1,"metrics_granularity":"1Minute","min_elb_capacity":null,"min_size":1,"mixed_instances_policy":[],"name_prefix":null,"placement_group":null,"protect_from_scale_in":false,"suspended_processes":null,"tag":[],"tags":null,"target_group_arns":null,"termination_policies":null,"timeouts":null,"wait_for_capacity_timeout":"10m","wait_for_elb_capacity":null},"after_unknown":{"arn":true,"availability_zones":[false],"default_cooldown":true,"health_check_type":true,"id":true,"initial_lifecycle_hook":[],"instance_refresh":[],"launch_template":[{"id":true,"name":true}],"mixed_instances_policy":[],"name":true,"service_linked_role_arn":true,"tag":[],"vpc_zone_identifier":true}}},{"address":"aws_launch_template.foobar","mode":"managed","type":"aws_launch_template","name":"foobar","provider_name":"registry.terraform.io/hashicorp/aws","change":{"actions":["create"],"before":null,"after":{"block_device_mappings":[],"capacity_reservation_specification":[],"cpu_options":[],"credit_specification":[],"description":null,"disable_api_termination":null,"ebs_optimized":null,"elastic_gpu_specifications":[],"elastic_inference_accelerator":[],"enclave_options":[],"hibernation_options":[],"iam_instance_profile":[],"image_id":"ami-1a2b3c","instance_initiated_shutdown_behavior":null,"instance_market_options":[],"instance_type":"t2.micro","kernel_id":null,"key_name":null,"license_specification":[],"monitoring":[],"name_prefix":"foobar","network_interfaces":[],"placement":[],"ram_disk_id":null,"security_group_names":null,"tag_specifications":[],"tags":null,"update_default_version":null,"user_data":null,"vpc_security_group_ids":null},"after_unknown":{"arn":true,"block_device_mappings":[],"capacity_reservation_specification":[],"cpu_options":[],"credit_specification":[],"default_version":true,"elastic_gpu_specifications":[],"elastic_inference_accelerator":[],"enclave_options":[],"hibernation_options":[],"iam_instance_profile":[],"id":true,"instance_market_options":[],"latest_version":true,"license_specification":[],"metadata_options":true,"monitoring":[],"name":true,"network_interfaces":[],"placement":[],"tag_specifications":[]}}}],"configuration":{"provider_config":{"aws":{"name":"aws","expressions":{"region":{"constant_value":"us-east-1"}}}},"root_module":{"resources":[{"address":"aws_autoscaling_group.bar","mode":"managed","type":"aws_autoscaling_group","name":"bar","provider_config_key":"aws","expressions":{"availability_zones":{"constant_value":["us-east-1a"]},"desired_capacity":{"constant_value":1},"health_check_grace_period":{"constant_value":600},"launch_template":[{"id":{"references":["aws_launch_template.foobar"]},"version":{"constant_value":"$Latest"}}],"max_size":{"constant_value":1},"min_size":{"constant_value":1}},"schema_version":0},{"address":"aws_launch_template.foobar","mode":"managed","type":"aws_launch_template","name":"foobar","provider_config_key":"aws","expressions":{"image_id":{"constant_value":"ami-1a2b3c"},"instance_type":{"constant_value":"t2.micro"},"name_prefix":{"constant_value":"foobar"}},"schema_version":0}]}}}
`)
)

func TestPolicyFind(t *testing.T) {
	var p policy
	assert.NoError(t, p.find(testTarget.Platform, testTarget.ResourceType))
	assert.Equal(t, ".reliably/policies/terraform/aws_autoscaling_group.rego", p.filepath)
	assert.Equal(t, "https://static.reliably.com/opa/terraform/aws_autoscaling_group.rego", p.uri)

	// check header
	headers := p.packageHeaders()
	assert.Len(t, headers, 3, "expected header length: 3")
	t.Logf("headers found --> %s", headers)
}

func TestEvaluate(t *testing.T) {
	var tfplan terraform.PlanRepresentation
	assert.NoError(t, json.Unmarshal(terraformTestPlan, &tfplan))
	// testTarget.Item = tfplan

	// get policy
	var p policy
	assert.NoError(t, p.find(testTarget.Platform, testTarget.ResourceType))

	// iterate module
	var targets []*Target
	for _, resource := range tfplan.PlannedValues.RootModule.Resources {
		targets = append(targets, &Target{
			ResourceType: resource.Type,
			Platform:     "terraform",
			Item:         resource,
		})
	}

	offendingTargets, err := p.evaluate(targets...)
	assert.NoError(t, err)
	assert.Len(t, offendingTargets, 1, "exoected offending items length 1")
	for _, target := range offendingTargets {
		for _, rule := range target.Result.Violations {
			t.Logf("Violation Detected:\n\tlevel: %d\n\tMessage: %s\n\tRuleID: %s\n\tRule Definition: %s\n\t----",
				rule.Level, rule.Message, rule.RuleID, rule.RuleDefinition,
			)
		}
	}
}
