package dns

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func RegisterIP(ctx context.Context, name, ip, hostedZone *string) error {
	r53api := route53.New(session.Must(session.NewSession()))
	_, err := r53api.ChangeResourceRecordSetsWithContext(ctx,
		&route53.ChangeResourceRecordSetsInput{
			HostedZoneId: hostedZone,
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: aws.String(route53.ChangeActionUpsert),
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name: name,
							ResourceRecords: []*route53.ResourceRecord{
								{
									Value: ip,
								},
							},
							TTL:  aws.Int64(60),
							Type: aws.String(route53.RRTypeA),
						},
					},
				},
			},
		})
	return err
}
