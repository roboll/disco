package disco

import "errors"

type AutoscaleProvider interface {
	GetInstances() ([]*string, error)
}

type AutoscaleConfig struct {
	GroupName string
}

type DNSProvider interface {
	SyncRecord(instances []*string) error
}

type DNSConfig struct {
	Domain string
	Port   string
	TTL    int64
	SSL    bool
}

const (
	RecordName    = "_etcd-server._tcp"
	SSLRecordName = "_etcd-server-ssl._tcp"
)

func (d *DNSConfig) Validate() error {
	if len(d.Domain) == 0 {
		return errors.New("Domain is required")
	}
	if len(d.Port) == 0 {
		return errors.New("Port is required")
	}
	if !(d.TTL > 0) {
		return errors.New("TTL must be > 0")
	}
	return nil
}

func (d *DNSConfig) GetRecordName() string {
	var record string
	if d.SSL {
		record = SSLRecordName + "." + d.Domain
	} else {
		record = RecordName + "." + d.Domain
	}
	return record
}
