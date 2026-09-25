package testimpl

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComposableComplete is the functional-test entrypoint. It runs under
// lib.RunSetupTestTeardown (apply -> assert -> destroy), so the mutating
// operation under test is the apply/destroy cycle itself; the assertions it
// delegates to are read-only cloud verification shared with the readonly
// suite.
func TestComposableComplete(t *testing.T, ctx types.TestContext) {
	verifyVpcReadOnly(t, ctx)
}

// TestComposableCompleteReadOnly is the readonly-test entrypoint. It runs
// under lib.RunNonDestructiveTest against already-deployed infrastructure
// and must not create, update, or delete anything -- it shares the same
// read-only verification as the functional suite, but on its own distinct
// implementation function per the TestComposable* naming lcaf-component-terratest
// requires for non-destructive runs.
func TestComposableCompleteReadOnly(t *testing.T, ctx types.TestContext) {
	verifyVpcReadOnly(t, ctx)
}

// verifyVpcReadOnly performs only read-only EC2 SDK calls (DescribeVpcs) and
// is safe to run against infrastructure that must not be mutated.
func verifyVpcReadOnly(t *testing.T, ctx types.TestContext) {
	awsClient := GetAWSEC2Client(t)

	t.Run("TestIsDeployed", func(t *testing.T) {
		vpcId := terraform.OutputContext(t, context.Background(), ctx.TerratestTerraformOptions(), "vpc_id")
		out, err := awsClient.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{
			VpcIds: []string{vpcId},
		})

		if err != nil {
			t.Errorf("Failure during DescribeVpcs: %v", err)
		}

		assert.Len(t, out.Vpcs, 1, "Expected VPC does not exists!")
	})

	t.Run("TestIsAvailable", func(t *testing.T) {
		vpcId := terraform.OutputContext(t, context.Background(), ctx.TerratestTerraformOptions(), "vpc_id")
		out, err := awsClient.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{
			VpcIds: []string{vpcId},
		})

		if err != nil {
			t.Errorf("Failure during DescribeVpcs: %v", err)
		}

		assert.Equal(t, "available", string(out.Vpcs[0].State), "VPC is not available!")
	})

	t.Run("TestCIDRBlock", func(t *testing.T) {
		vpcId := terraform.OutputContext(t, context.Background(), ctx.TerratestTerraformOptions(), "vpc_id")
		vpcCidrBlock := terraform.OutputContext(t, context.Background(), ctx.TerratestTerraformOptions(), "vpc_cidr_block")
		out, err := awsClient.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{
			VpcIds: []string{vpcId},
		})

		if err != nil {
			t.Errorf("Failure during DescribeVpcs: %v", err)
		}

		assert.Equal(t, vpcCidrBlock, *out.Vpcs[0].CidrBlock, "Expected VPC CIDR Block does not match current VPC CIDR Block!")
	})

	t.Run("TestTags", func(t *testing.T) {
		vpcId := terraform.OutputContext(t, context.Background(), ctx.TerratestTerraformOptions(), "vpc_id")
		tags := terraform.OutputMapContext(t, context.Background(), ctx.TerratestTerraformOptions(), "vpc_tags")
		out, err := awsClient.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{
			VpcIds: []string{vpcId},
		})

		if err != nil {
			t.Errorf("Failure during DescribeVpcs: %v", err)
		}

		found := 0

		for tagk, tagv := range tags {
			for _, tag2 := range out.Vpcs[0].Tags {
				if tagk == *tag2.Key && tagv == *tag2.Value {
					found++
					break
				}
			}
		}

		assert.GreaterOrEqual(t, found, len(tags), "Expected VPC tags does not match current VPC tags!")
	})
}

func GetAWSEC2Client(t *testing.T) *ec2.Client {
	awsEc2Client := ec2.NewFromConfig(GetAWSConfig(t))
	return awsEc2Client
}

func GetAWSConfig(t *testing.T) (cfg aws.Config) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	require.NoErrorf(t, err, "unable to load SDK config, %v", err)
	return cfg
}
