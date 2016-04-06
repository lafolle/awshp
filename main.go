package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/lafolle/goodhosts"
)

var (
	region  string
	stackId string
)

func init() {
	flag.StringVar(&stackId, "stackId", "", "stackId")
	flag.StringVar(&region, "region", "", "region")
}

func populateHostsFile(instanceMap map[*string]map[*string]*string) {
	hosts, err := goodhosts.NewHosts()
	if err != nil {
		fmt.Println("failed to get hosts")
		return
	}
	for _, imap := range instanceMap {
		for hostname, ip := range imap {
			if hosts.Has(*ip, *hostname) {
				fmt.Println("host already has ", *ip, *hostname)
				continue
			}
			hosts.Add(*ip, *hostname, "")
		}
	}
	if err := hosts.Flush(); err != nil {
		fmt.Println("failed to write to file", err)
	}
}

func main() {
	flag.Parse()
	if region == "" || stackId == "" {
		flag.PrintDefaults()
		return
	}
	var instanceMap map[*string]map[*string]*string = make(map[*string]map[*string]*string)
	opsWorksInstance := opsworks.New(session.New(&aws.Config{Region: aws.String(region)}))
	if opsWorksInstance == nil {
		fmt.Println("Failed to get opsworks instance.")
		return
	}
	describeLayersOutput, err := opsWorksInstance.DescribeLayers(&opsworks.DescribeLayersInput{
		StackId: aws.String(stackId),
	})
	if err != nil {
		fmt.Println("could not connect fetch layer info", err)
		return
	}
	for _, layer := range describeLayersOutput.Layers {
		describeInstancesOutput, err := opsWorksInstance.DescribeInstances(&opsworks.DescribeInstancesInput{
			LayerId: layer.LayerId,
		})
		if err != nil {
			fmt.Println("failed to get instances for layer: ", layer.LayerId, err)
			continue
		}
		instanceMap[layer.LayerId] = make(map[*string]*string)
		for _, instance := range describeInstancesOutput.Instances {
			if *instance.Status != *aws.String("online") {
				fmt.Println("instance is offline (skipping)", *instance.Hostname)
				continue
			} else {
				fmt.Println("instance is online", *instance.Hostname)
			}
			instanceMap[layer.LayerId][instance.Hostname] = instance.PublicIp
		}
	}
	populateHostsFile(instanceMap)
}
