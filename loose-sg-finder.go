package main

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/exp/slices"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
)

func main() {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	var used_sgs []string
	ec2_client := ec2.NewFromConfig(cfg)
	all_sg, err := ec2_client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		log.Fatal(err)
	}
	// Taking Care of EC2 & Lambdas
	network_interfaces, err := ec2_client.DescribeNetworkInterfaces(context.TODO(), &ec2.DescribeNetworkInterfacesInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, element := range network_interfaces.NetworkInterfaces {
		for _, sg := range element.Groups {
			used_sgs = append(used_sgs, *sg.GroupId)
		}
	}
	// Taking Care of RDS instances
	rdsClient := rds.NewFromConfig(cfg)
	rdses, err := rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, rds_instance := range rdses.DBInstances {
		for _, rds_sg := range rds_instance.VpcSecurityGroups {
			used_sgs = append(used_sgs, *rds_sg.VpcSecurityGroupId)
		}
	}
	// Taking Care of Classic LB
	elbClient := elasticloadbalancing.NewFromConfig(cfg)
	classicElbs, err := elbClient.DescribeLoadBalancers(context.TODO(), &elasticloadbalancing.DescribeLoadBalancersInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, elb := range classicElbs.LoadBalancerDescriptions {
		for _, elbSg := range elb.SecurityGroups {
			used_sgs = append(used_sgs, elbSg)
		}
	}

	// Taking Care of App LBs
	elbv2Client := elasticloadbalancingv2.NewFromConfig(cfg)
	appElbs, err := elbv2Client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, alb := range appElbs.LoadBalancers {
		for _, albSg := range alb.SecurityGroups {
			used_sgs = append(used_sgs, albSg)
		}
	}

	// Taking care of Elasticache
	elastiCacheClient := elasticache.NewFromConfig(cfg)
	elastiCacheInstances, err := elastiCacheClient.DescribeCacheClusters(context.TODO(), &elasticache.DescribeCacheClustersInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, cacheCluster := range elastiCacheInstances.CacheClusters {
		for _, cacheSg := range cacheCluster.SecurityGroups {
			used_sgs = append(used_sgs, *cacheSg.SecurityGroupId)
		}
	}

	// Taking Care of RedShift
	redshiftClient := redshift.NewFromConfig(cfg)
	redshiftClusters, err := redshiftClient.DescribeClusters(context.TODO(), &redshift.DescribeClustersInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, cacheCluster := range redshiftClusters.Clusters {
		for _, cacheSg := range cacheCluster.VpcSecurityGroups {
			used_sgs = append(used_sgs, *cacheSg.VpcSecurityGroupId)
		}
	}

	for _, element := range all_sg.SecurityGroups {
		if !slices.Contains(used_sgs, *element.GroupId) {
			fmt.Printf("%s\t%s\t%s\n", *element.GroupId, *element.GroupName, *element.Description)
		}
	}
}
