package main

import (
	"flag"
	"log"
	"os"

	"github.com/kitkitcode/disco/disco"
)

var zone string
var domain string
var ssl bool
var port string
var ttl int64
var wait bool

var region string
var group string
var valueType string

func init() {
	flag.StringVar(&zone, "zone", "", "hosted zone id")
	flag.StringVar(&domain, "domain", "etcd.local", "domain name")
	flag.BoolVar(&ssl, "ssl", false, "use ssl")
	flag.StringVar(&port, "port", "2380", "port")
	flag.Int64Var(&ttl, "ttl", 60, "record ttl")
	flag.BoolVar(&wait, "wait", false, "wait for propagation confirmation")

	flag.StringVar(&region, "region", "", "aws region")
	flag.StringVar(&group, "group", "", "autoscaling group")
	flag.StringVar(&valueType, "value-type", "", "value type - one of private-ip, public-ip, private-dns, public-dns")
}

func main() {
	flag.Parse()

	as, err := disco.AmazonProvider(&disco.AmazonConfig{
		Region:    region,
		GroupName: group,
		ValueType: valueType,
	})
	if err != nil {
		log.Fatal(err)
	}

	instances, err := as.GetInstances()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("got %d instances\n", len(instances))

	dnsc := &disco.DNSConfig{
		Domain: domain,
		Port:   port,
		TTL:    ttl,
		SSL:    ssl,
	}
	dnsr := &disco.Route53Config{
		ZoneID: zone,
		Wait:   wait,
	}

	dns, err := disco.Route53Provider(dnsc, dnsr)
	if err != nil {
		log.Fatal(err)
	}
	err = dns.SyncRecord(instances)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
