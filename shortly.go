package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapprunner"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ShortlyStackProps struct {
	awscdk.StackProps
}

func NewShortlyStack(scope constructs.Construct, id string, props *ShortlyStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	const project = "shortly"

	resourceName := func(resourceType string) string {
		return project + "-" + resourceType
	}

	// Database DynamoDB
	// Shortcut table
	shortcutTableName := resourceName("shortcut")
	shortcutTable := awsdynamodb.NewTableV2(stack, jsii.String(shortcutTableName), &awsdynamodb.TablePropsV2{
		TableName: jsii.String(shortcutTableName),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	// ECR
	imageName := resourceName("image")
	image := awsecrassets.NewDockerImageAsset(
		stack, jsii.String(imageName),
		&awsecrassets.DockerImageAssetProps{
			AssetName: jsii.String(imageName),
			Directory: jsii.String("app"),
			Platform:  awsecrassets.Platform_LINUX_AMD64(),
		},
	)

	// IAM Role to access ECR
	roleName := resourceName("ecr-access-role")
	role := awsiam.NewRole(stack, jsii.String(roleName), &awsiam.RoleProps{
		RoleName: jsii.String(roleName),
		AssumedBy: awsiam.NewServicePrincipal(
			jsii.String("build.apprunner.amazonaws.com"), nil,
		),
	})
	image.Repository().GrantRead(role)
	image.Repository().GrantPull(role)

	// Instance role
	instanceRoleName := resourceName("instance-role")
	instanceRole := awsiam.NewRole(stack, jsii.String(instanceRoleName), &awsiam.RoleProps{
		RoleName: jsii.String(instanceRoleName),
		AssumedBy: awsiam.NewServicePrincipal(
			jsii.String("tasks.apprunner.amazonaws.com"), nil,
		),
	})

	// Grant access to DynamoDB
	shortcutTable.GrantReadWriteData(instanceRole)

	// Service
	serviceName := resourceName("service")
	service := awsapprunner.NewCfnService(stack, jsii.String(serviceName), &awsapprunner.CfnServiceProps{
		ServiceName: jsii.String(serviceName),
		InstanceConfiguration: &awsapprunner.CfnService_InstanceConfigurationProperty{
			InstanceRoleArn: instanceRole.RoleArn(),
		},
		SourceConfiguration: &awsapprunner.CfnService_SourceConfigurationProperty{
			AuthenticationConfiguration: &awsapprunner.CfnService_AuthenticationConfigurationProperty{
				AccessRoleArn: role.RoleArn(),
			},
			ImageRepository: &awsapprunner.CfnService_ImageRepositoryProperty{
				ImageIdentifier:     image.ImageUri(),
				ImageRepositoryType: jsii.String("ECR"),
				ImageConfiguration: &awsapprunner.CfnService_ImageConfigurationProperty{
					Port: jsii.String("8080"),
					RuntimeEnvironmentVariables: []interface{}{
						&awsapprunner.CfnService_KeyValuePairProperty{
							Name:  jsii.String("BaseURL"),
							Value: jsii.String(os.Getenv("BaseURL")),
						},
						&awsapprunner.CfnService_KeyValuePairProperty{
							Name:  jsii.String("ShortcutsTableName"),
							Value: shortcutTable.TableName(),
						},
					},
				},
			},
		},
	})

	// Output
	awscdk.NewCfnOutput(stack, jsii.String("ServiceURL"), &awscdk.CfnOutputProps{
		ExportName: jsii.String("ServiceURL"),
		Value:      service.AttrServiceUrl(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewShortlyStack(app, "ShortlyStack", &ShortlyStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
