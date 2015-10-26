# disco

_for etcd discovery_

## about

_disco_ is a tool to help with etcd dns discovery using autoscaling groups. Run _disco_ on startup of machines in the group and it will update dns srv records to point to the instances to allow for dns discovery.

It currently supports aws autoscaling and route53 dns, but should be trivial to extend for other providers. See [providers](disco/provider.go).

## usage

_disco_ accepts arguments for all etcd configuration parameters, and expects that hosts should be configured identically - particularly, that it uses the same port and client/peer settings on every host.

```
  -domain string
        domain name (default "etcd.local")
  -group string
        autoscaling group
  -port string
        port (default "2380")
  -region string
        aws region
  -ssl
        use ssl
  -ttl int
        record ttl (default 60)
  -value-type string
        value type - one of private-ip, public-ip, private-dns, public-dns
  -wait
        wait for propagation confirmation
  -zone string
        hosted zone id
```

### aws usage

_disco_ expects iam permissions to be in place, and relies on the default aws credentials chaining (it will take advantage of iam credentials automatically).
