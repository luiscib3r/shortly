package main

import (
	"os"
	"reflect"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapplicationautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
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

	// Cluster VPC
	vpcName := resourceName("vpc")
	vpc := awsec2.NewVpc(stack, jsii.String(vpcName), &awsec2.VpcProps{
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String("10.0.0.0/16")),
		NatGateways: jsii.Number(1),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("public"),
				CidrMask:   jsii.Number(24),
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
			{
				Name:       jsii.String("private"),
				CidrMask:   jsii.Number(24),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
			},
		},
	})

	// Private Subnets
	privateSubNets := vpc.PrivateSubnets()
	var privateSubNetIds []string
	for _, subnet := range *privateSubNets {
		privateSubNetIds = append(privateSubNetIds, *subnet.SubnetId())
	}

	// Redis Subnet Group
	redisSubNetGroup := awselasticache.NewCfnSubnetGroup(
		stack,
		jsii.String(resourceName("redis-subnet-group")),
		&awselasticache.CfnSubnetGroupProps{
			Description: jsii.String("Redis Subnet Group"),
			SubnetIds:   jsii.Strings(privateSubNetIds...),
		},
	)

	// Security Groups
	// Fargate Security Group
	fargateSecurityGroupName := resourceName("fargate-sg")
	fargateSecurityGroup := awsec2.NewSecurityGroup(stack, jsii.String(fargateSecurityGroupName), &awsec2.SecurityGroupProps{
		SecurityGroupName: jsii.String(fargateSecurityGroupName),
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(true),
	})

	fargateSecurityGroup.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(8080)),
		jsii.String("Allow inbound traffic on port 8080"),
		nil,
	)

	// Redis Security Group
	redisSecurityGroupName := resourceName("redis-sg")
	redisSecurityGroup := awsec2.NewSecurityGroup(stack, jsii.String(redisSecurityGroupName), &awsec2.SecurityGroupProps{
		SecurityGroupName: jsii.String(redisSecurityGroupName),
		Vpc:               vpc,
		AllowAllOutbound:  jsii.Bool(true),
	})

	redisSecurityGroup.AddIngressRule(
		fargateSecurityGroup,
		awsec2.Port_Tcp(jsii.Number(6379)),
		jsii.String("Allow inbound traffic on port 6379 from Fargate Service"),
		nil,
	)

	// Database Redis
	redisName := resourceName("redis")
	redisCluster := awselasticache.NewCfnCacheCluster(
		stack,
		jsii.String(redisName),
		&awselasticache.CfnCacheClusterProps{
			ClusterName:          jsii.String(redisName),
			Engine:               jsii.String("redis"),
			CacheNodeType:        jsii.String("cache.t3.micro"),
			NumCacheNodes:        jsii.Number(1),
			CacheSubnetGroupName: redisSubNetGroup.Ref(),
			VpcSecurityGroupIds:  &[]*string{redisSecurityGroup.SecurityGroupId()},
		},
	)

	redisEndpoint := *redisCluster.AttrRedisEndpointAddress() + ":" + *redisCluster.AttrRedisEndpointPort()

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
				"BaseURL":            jsii.String("https://" + domainName),
				"ShortcutsTableName": shortcutTable.TableName(),
				"RedisEndpoint":      jsii.String(redisEndpoint),
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

	// Cluster with VPC
	clusterName := resourceName("cluster")
	cluster := awsecs.NewCluster(stack, jsii.String(clusterName), &awsecs.ClusterProps{
		ClusterName: jsii.String(clusterName),
		Vpc:         vpc,
	})

	securityGroups := &[]awsec2.ISecurityGroup{fargateSecurityGroup}

	// Fargate Load Balanced Service
	serviceName := resourceName("service")
	service := awsecspatterns.NewApplicationLoadBalancedFargateService(stack,
		jsii.String(serviceName),
		&awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
			ServiceName:        jsii.String(serviceName),
			LoadBalancerName:   jsii.String(resourceName("lb")),
			Cluster:            cluster,
			SecurityGroups:     securityGroups,
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
