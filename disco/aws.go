package disco

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type amazon struct {
	c *AmazonConfig
}

type AmazonConfig struct {
	Region    string
	GroupName string
	ValueType string
}

type ValueType string

const (
	TypePrivateIP       = "private-ip"
	TypePublicIP        = "public-ip"
	TypePrivateHostname = "private-dns"
	TypePublicHostname  = "public-dns"
)

func AmazonProvider(c *AmazonConfig) (AutoscaleProvider, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}
	return &amazon{c: c}, nil
}

func (c *AmazonConfig) validate() error {
	if len(c.GroupName) == 0 {
		return errors.New("GroupName is required")
	}
	return nil
}

func (p *amazon) GetInstances() ([]*string, error) {
	region, err := p.getRegion()
	if err != nil {
		return nil, err
	}

	as := autoscaling.New(&aws.Config{
		Region: &region,
	})
	ec := ec2.New(&aws.Config{
		Region: &region,
	})

	log.Printf("Searching groups with name: %s\n", p.c.GroupName)
	output, err := as.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&p.c.GroupName},
	})
	if err != nil {
		return nil, err
	}
	if len(output.AutoScalingGroups) != 1 {
		return nil, errors.New("Expected 1 autoscaling group for name.")
	}

	instances := []*string{}
	for _, group := range output.AutoScalingGroups {
		for _, instance := range group.Instances {
			out, err := ec.DescribeInstances(&ec2.DescribeInstancesInput{
				InstanceIds: []*string{instance.InstanceId},
			})
			if err != nil {
				return nil, err
			}
			for _, res := range out.Reservations {
				for _, i := range res.Instances {
					switch p.c.ValueType {
					case TypePrivateIP:
						instances = append(instances, i.PrivateIpAddress)
					case TypePublicIP:
						instances = append(instances, i.PublicIpAddress)
					case TypePrivateHostname:
						instances = append(instances, i.PrivateDnsName)
					case TypePublicHostname:
						instances = append(instances, i.PublicDnsName)
					default:
						return nil, errors.New("must specify a valid value type field")
					}
				}
			}
		}
	}
	return instances, nil
}

func (p *amazon) getRegion() (string, error) {
	if len(p.c.Region) == 0 {
		meta := ec2metadata.New(&ec2metadata.Config{})
		return meta.Region()
	}
	return p.c.Region, nil
}

func (p *amazon) getLocalInstanceID() (string, error) {
	meta := ec2metadata.New(&ec2metadata.Config{})
	return meta.GetMetadata("instance-id")
}

func (p *amazon) getLocalInstanceAutoscaleGroup() (*string, error) {
	id, err := p.getLocalInstanceID()
	if err != nil {
		return nil, fmt.Errorf("Unable to get instance id from metadata service. %s", err)
	}

	region, err := p.getRegion()
	if err != nil {
		return nil, err
	}

	ec := ec2.New(&aws.Config{Region: &region})
	output, err := ec.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{&id},
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to get instance data from ec2. %s", err)
	}

	if len(output.Reservations) != 1 {
		return nil, errors.New("Expected one reservation when searching by instance-id.")
	}
	for _, res := range output.Reservations {
		if len(res.Instances) != 1 {
			return nil, errors.New("Expected one instance when searching by instance-id.")
		}
		for _, instance := range res.Instances {
			for _, tag := range instance.Tags {
				if "aws:autoscaling:groupName" == *tag.Key {
					return tag.Value, nil
				}
			}
		}
	}
	return nil, errors.New("Unable to find autoscaling group tag.")
}
