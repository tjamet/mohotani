package route53

import (
	"fmt"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Route53 struct {
	client *route53.Route53
}

func NewRoute53() *Route53 {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
	}))

	svc := route53.New(sess)
	return &Route53{svc}
}

func (r53 *Route53) Update(domain string, targets ...string) error {
	if len(targets) == 0 {
		return fmt.Errorf("no target provided, abording")
	}
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	isCNAME := net.ParseIP(targets[0]) == nil
	for _, target := range targets[1:] {
		if isCNAME != (net.ParseIP(target) == nil) {
			return fmt.Errorf("mixed targets between IP and CNAMES")
		}
	}
	if isCNAME && len(targets) != 1 {
		return fmt.Errorf("cannot set CNAME to multiple domains %v", targets)
	}
	t := "CNAME"
	if !isCNAME {
		t = "A"
	}
	records := []*route53.ResourceRecord{}
	for _, tgt := range targets {
		records = append(records,
			&route53.ResourceRecord{ // Required
				Value: aws.String(tgt), // Required
			},
		)
	}

	zones, err := r53.client.ListHostedZones(&route53.ListHostedZonesInput{})
	if err != nil {
		return err
	}
	for _, zone := range zones.HostedZones {
		if strings.HasSuffix(domain, *zone.Name) {
			params := &route53.ChangeResourceRecordSetsInput{
				ChangeBatch: &route53.ChangeBatch{ // Required
					Changes: []*route53.Change{ // Required
						{ // Required
							Action: aws.String("UPSERT"), // Required
							ResourceRecordSet: &route53.ResourceRecordSet{ // Required
								Name:            aws.String(domain), // Required
								Type:            aws.String(t),      // Required
								ResourceRecords: records,
								TTL:             aws.Int64(60),
								Weight:          aws.Int64(100),
								SetIdentifier:   aws.String("Updated by mohotani"),
							},
						},
					},
					Comment: aws.String("Updated by mohotani"),
				},
				HostedZoneId: zone.Id, // Required
			}
			_, err := r53.client.ChangeResourceRecordSets(params)
			return err
		}
	}
	return fmt.Errorf("unknown zone for host %s", domain)
}
