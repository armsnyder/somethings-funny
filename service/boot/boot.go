package boot

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/sirupsen/logrus"
	"net/http"
	"somethings-funny/service/dns"
)

func RegisterOwnDNS(serviceDomain, hostedZone *string) error {
	logrus.Info("Discovering own IP address")
	ownIp, err := discoverIP()
	if err != nil {
		return err
	}
	logrus.Infof("Registering own IP address %s with DNS", *ownIp)
	err = dns.RegisterIP(context.Background(), serviceDomain, ownIp, hostedZone)
	if err != nil {
		return err
	}
	return nil
}

func discoverIP() (*string, error) {
	sess := session.Must(session.NewSession())
	ecsapi := ecs.New(sess)
	ec2api := ec2.New(sess)
	_ = ec2.New(sess)
	resp, err := http.Get("http://169.254.170.2/v2/metadata")
	if err != nil {
		return nil, err
	}
	parsedBody := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&parsedBody); err != nil {
		return nil, err
	}
	taskArn := parsedBody["TaskARN"].(string)
	eniId, err := getEniId(ecsapi, &taskArn)
	if err != nil {
		return nil, err
	}
	output, err := ecsapi.DescribeTasks(&ecs.DescribeTasksInput{
		Tasks: []*string{&taskArn},
	})
	if err != nil {
		return nil, err
	}

	for _, a := range output.Tasks[0].Attachments {
		if *a.Type == "networkInterfaceId" {
			eniId = a.Id
		}
	}
	output1, err := ec2api.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{eniId},
	})
	if err != nil {
		return nil, err
	}
	return output1.NetworkInterfaces[0].Association.PublicIp, nil
}

func getEniId(ecsapi *ecs.ECS, taskArn *string) (*string, error) {
	output, err := ecsapi.DescribeTasks(&ecs.DescribeTasksInput{
		Tasks: []*string{taskArn},
	})
	if err != nil {
		return nil, err
	}
	var eniId *string
	for _, a := range output.Tasks[0].Attachments {
		for _, d := range a.Details {
			if *d.Name == "networkInterfaceId" {
				eniId = d.Value
			}
		}
	}
	return eniId, nil
}
