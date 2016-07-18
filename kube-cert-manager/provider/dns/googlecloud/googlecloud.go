package googlecloud

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
)

type DNSProvider struct {
	client  *dns.Service
	project string
}

func NewDNSProvider(project, serviceAccount string) (*DNSProvider, error) {
	jsonKey, err := ioutil.ReadFile(serviceAccount)
	if err != nil {
		return nil, err
	}

	jwtConfig, err := google.JWTConfigFromJSON(jsonKey, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, err
	}

	client, err := dns.New(jwtConfig.Client(context.Background()))
	if err != nil {
		return nil, err
	}

	return &DNSProvider{client, project}, nil
}

func (p *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	zone, err := p.getHostedZone(domain)
	if err != nil {
		return err
	}

	record := &dns.ResourceRecordSet{
		Name:    fqdn,
		Rrdatas: []string{value},
		Ttl:     int64(ttl),
		Type:    "TXT",
	}

	change := &dns.Change{
		Additions: []*dns.ResourceRecordSet{record},
	}

	changesCreateCall, err := p.client.Changes.Create(p.project, zone, change).Do()
	if err != nil {
		return err
	}

	for changesCreateCall.Status == "pending" {
		time.Sleep(time.Second)
		changesCreateCall, err = p.client.Changes.Get(p.project, zone, changesCreateCall.Id).Do()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := p.getHostedZone(domain)
	if err != nil {
		return err
	}

	records, err := p.findTxtRecords(zone, fqdn)
	if err != nil {
		return err
	}

	for _, record := range records {
		change := &dns.Change{
			Deletions: []*dns.ResourceRecordSet{record},
		}
		_, err = p.client.Changes.Create(p.project, zone, change).Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 300 * time.Second, 5 * time.Second
}

func (p *DNSProvider) getHostedZone(domain string) (string, error) {
	zones, err := p.client.ManagedZones.List(p.project).Do()
	if err != nil {
		return "", err
	}

	for _, zone := range zones.ManagedZones {
		if strings.HasSuffix(domain+".", zone.DnsName) {
			return zone.Name, nil
		}
	}

	return "", fmt.Errorf("No matching GoogleCloud domain found for domain %s", domain)
}

func (p *DNSProvider) findTxtRecords(zone, fqdn string) ([]*dns.ResourceRecordSet, error) {
	records, err := p.client.ResourceRecordSets.List(p.project, zone).Do()
	if err != nil {
		return nil, err
	}

	found := []*dns.ResourceRecordSet{}
	for _, record := range records.Rrsets {
		if record.Type == "TXT" && record.Name == fqdn {
			found = append(found, record)
		}
	}
	return found, nil
}
