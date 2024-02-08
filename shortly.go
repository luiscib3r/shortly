package main

import (
	"os"
	"reflect"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapplicationautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
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

	domainNameVar := stack.Node().TryGetContext(jsii.String(id + ":domainName"))
	hostedZoneNameVar := stack.Node().TryGetContext(jsii.String(id + ":hostedZoneName"))

	domainName := reflect.ValueOf(domainNameVar).String()
	hostedZoneName := reflect.ValueOf(hostedZoneNameVar).String()

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

	// ECS Task Role
	taskRoleName := resourceName("task-role")
	taskRole := awsiam.NewRole(stack, jsii.String(taskRoleName), &awsiam.RoleProps{
		RoleName: jsii.String(taskRoleName),
		AssumedBy: awsiam.NewServicePrincipal(
			jsii.String("ecs-tasks.amazonaws.com"), nil,
		),
	})

	// Grant access to DynamoDB
	shortcutTable.GrantReadWriteData(taskRole)

	// ECR
	imageName := resourceName("image")
	image := awsecrassets.NewDockerImageAsset(
		stack, jsii.String(imageName),
		&awsecrassets.DockerImageAssetProps{
			AssetName: jsii.String(imageName),
			Directory: jsii.String("app"),
			Platform:  awsecrassets.Platform_LINUX_ARM64(),
		},
	)

	// Fargate Task Definition
	taskDefName := resourceName("task-def")
	taskDef := awsecs.NewFargateTaskDefinition(stack, jsii.String(taskDefName), &awsecs.FargateTaskDefinitionProps{
		TaskRole: taskRole,
		RuntimePlatform: &awsecs.RuntimePlatform{
			CpuArchitecture: awsecs.CpuArchitecture_ARM64(),
		},
		Cpu:            jsii.Number(512),
		MemoryLimitMiB: jsii.Number(1024),
	})

	// Logging
	logging := awsecs.LogDrivers_AwsLogs(&awsecs.AwsLogDriverProps{
		StreamPrefix: jsii.String("shortly"),
	})

	// Task Container
	containerName := resourceName("container")

	taskDef.AddContainer(
		jsii.String(containerName),
		&awsecs.ContainerDefinitionOptions{
			Image:   awsecs.ContainerImage_FromDockerImageAsset(image),
			Logging: logging,
			PortMappings: &[]*awsecs.PortMapping{
				{
					ContainerPort: jsii.Number(8080),
				},
			},
			Environment: &map[string]*string{
				"BaseURL":            jsii.String(os.Getenv("BaseURL")),
				"ShortcutsTableName": shortcutTable.TableName(),
			},
			HealthCheck: &awsecs.HealthCheck{
				Command: jsii.Strings("CMD-SHELL", "curl -f http://localhost:8080/healthcheck || exit 1"),
			},
		},
	)

	// Hosted Zone
	hostedZone := awsroute53.HostedZone_FromLookup(
		stack, jsii.String(resourceName("hosted-zone")),
		&awsroute53.HostedZoneProviderProps{
			DomainName: jsii.String(hostedZoneName),
		},
	)

	// Fargate Load Balanced Service
	serviceName := resourceName("service")
	service := awsecspatterns.NewApplicationLoadBalancedFargateService(stack,
		jsii.String(serviceName),
		&awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
			ServiceName:        jsii.String(serviceName),
			LoadBalancerName:   jsii.String(resourceName("lb")),
			TaskDefinition:     taskDef,
			PublicLoadBalancer: jsii.Bool(true),
			DesiredCount:       jsii.Number(1),
			RedirectHTTP:       jsii.Bool(true),
			Protocol:           awselasticloadbalancingv2.ApplicationProtocol_HTTPS,
			TargetProtocol:     awselasticloadbalancingv2.ApplicationProtocol_HTTP,
			DomainName:         jsii.String(domainName),
			DomainZone:         hostedZone,
		},
	)

	// Health Check
	albHealthCheck := &awselasticloadbalancingv2.HealthCheck{
		Path:     jsii.String("/healthcheck"),
		Port:     jsii.String("8080"),
		Protocol: awselasticloadbalancingv2.Protocol_HTTP,
	}

	service.TargetGroup().ConfigureHealthCheck(albHealthCheck)

	// Auto scaling
	scaling := service.Service().AutoScaleTaskCount(&awsapplicationautoscaling.EnableScalingProps{
		MaxCapacity: jsii.Number(100),
	})

	scaling.ScaleOnCpuUtilization(
		jsii.String(resourceName("cpu-scaling")),
		&awsecs.CpuUtilizationScalingProps{
			TargetUtilizationPercent: jsii.Number(90),
			ScaleInCooldown:          awscdk.Duration_Seconds(jsii.Number(60)),
			ScaleOutCooldown:         awscdk.Duration_Seconds(jsii.Number(60)),
		},
	)

	scaling.ScaleOnMemoryUtilization(
		jsii.String(resourceName("memory-scaling")),
		&awsecs.MemoryUtilizationScalingProps{
			TargetUtilizationPercent: jsii.Number(90),
			ScaleInCooldown:          awscdk.Duration_Seconds(jsii.Number(60)),
			ScaleOutCooldown:         awscdk.Duration_Seconds(jsii.Number(60)),
		},
	)

	// Output
	awscdk.NewCfnOutput(stack, jsii.String("ServiceURL"), &awscdk.CfnOutputProps{
		Value: service.LoadBalancer().LoadBalancerDnsName(),
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
	// return nil

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
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
