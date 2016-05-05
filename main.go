package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/lafolle/etchosts"
	"net"
	"os"
)

var (
	region  string
	stackId string
	dryRun  bool
)

type Instance struct {
	Status   *string
	Hostname *string
	PublicIp *string
}

func init() {
	flag.StringVar(&stackId, "stackId", "", "stackId")
	flag.StringVar(&region, "region", "", "region")
	flag.BoolVar(&dryRun, "dryRun", false, "do not commit changes")
}

func populateHostsFile(instanceMap map[string]Instance) error {
	hosts, err := etchosts.New("")
	if err != nil {
		return err
	}
	for hostname, instance := range instanceMap {
		entry, err := hosts.Read(hostname)
		if err == nil {
			if instance.PublicIp != nil && *instance.PublicIp != entry.Ipaddr.String() {
				fmt.Println("action: update")
				if err := hosts.Update(etchosts.Entry{
					Hostname: hostname,
					Ipaddr:   net.ParseIP(*instance.PublicIp),
				}); err != nil {
					fmt.Println("fail: action update", err)
				}
			}
			if instance.PublicIp == nil {
				fmt.Println("action: delete")
				if err := hosts.Delete(hostname); err != nil {
					fmt.Println("fail: action delete", hostname)
				}
			}
		} else if instance.PublicIp != nil {
			fmt.Println("action: create")
			if err := hosts.Create(etchosts.Entry{
				Hostname: hostname,
				Ipaddr:   net.ParseIP(*instance.PublicIp),
			}); err != nil {
				fmt.Println("fail: action create", err)
			}
		}
	}
	if dryRun {
		fmt.Println(hosts)
	} else {
		if err := hosts.Flush(); err != nil {
			fmt.Println("failed to write to file", err)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	if region == "" || stackId == "" {
		flag.PrintDefaults()
		return
	}
	fmt.Println("dryRun:", dryRun)
	var instanceMap map[string]Instance = make(map[string]Instance)
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
		for _, instance := range describeInstancesOutput.Instances {
			instanceMap[*instance.Hostname] = Instance{
				Status:   instance.Status,
				Hostname: instance.Hostname,
				PublicIp: instance.PublicIp,
			}
		}
	}
	if err := populateHostsFile(instanceMap); err != nil {
		fmt.Println("fail to populate etchosts:", err)
		os.Exit(1)
	}
}
