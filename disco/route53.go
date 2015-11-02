package disco

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

type rte53 struct {
	c *DNSConfig
	r *Route53Config
}

type Route53Config struct {
	ZoneID string
	Wait   bool
}

const (
	resourceTemplate = "0 0 %s %s"
)

var (
	upsert  = route53.ChangeActionUpsert
	srv     = route53.RRTypeSrv
	comment = "Managed by disco: https://github.com/kitkitcode/disco"
)

func Route53Provider(c *DNSConfig, r *Route53Config) (DNSProvider, error) {
	err := r.validate()
	if err != nil {
		return nil, err
	}

	return &rte53{
		c: c,
		r: r,
	}, nil
}

func (c *Route53Config) validate() error {
	if len(c.ZoneID) == 0 {
		return errors.New("ZoneID is required")
	}
	return nil
}

func (p *rte53) SyncRecord(instances []*string) error {
	client := route53.New(&aws.Config{})

	name := p.c.GetRecordName()

	records := []*route53.ResourceRecord{}
	for _, i := range instances {
		val := fmt.Sprintf(resourceTemplate, p.c.Port, *i)
		records = append(records, &route53.ResourceRecord{Value: &val})
	}

	change := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &p.r.ZoneID,
		ChangeBatch: &route53.ChangeBatch{
			Comment: &comment,
			Changes: []*route53.Change{
				{
					Action: &upsert,
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            &name,
						TTL:             &p.c.TTL,
						Type:            &srv,
						ResourceRecords: records,
					},
				},
			},
		},
	}

	result, err := client.ChangeResourceRecordSets(&change)
	if err != nil {
		log.Printf("got an error, will retry in 10s. %s", err)
		time.Sleep(10 * time.Second)
		result, err = client.ChangeResourceRecordSets(&change)
		if err != nil {
			log.Fatalf("received second error, failing. %s", err)
		}
	}

	log.Println("dns change submitted")

	if p.r.Wait {
		log.Println("waiting for change confirmation")
		for {
			change, err := client.GetChange(&route53.GetChangeInput{
				Id: result.ChangeInfo.Id,
			})
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("status is %s\n", *change.ChangeInfo.Status)
			if *change.ChangeInfo.Status == route53.ChangeStatusPending {
				log.Println("sleeping for 10s")
				time.Sleep(10 * time.Second)
			} else {
				return nil
			}
		}
	} else {
		log.Println("not waiting for change confirmation")
	}

	return nil
}
