package modules

import (
	"domain0/models"
)

type DnsObj interface {
	Create() error
	Update() error
	Delete() error
}

type DnsObjList interface {
	GetDNSList(d *models.Domain) error
	MultipleSelectWithIds(ids []string, r *interface{}) error
}

func DnsObjGen(d *models.Domain) DnsObj {
	if d.Vendor == "cloudflare" {
		return &CloudflareDNS{}
	}

	if d.Vendor == "dnspod" {
		return &TencentDNS{}
	}

	if d.Vendor == "aliyun" {
		return &AliDNS{}
	}

	return nil
}

func DnsListObjGen(d *models.Domain) DnsObjList {
	if d.Vendor == "cloudflare" {
		return &CloudflareDNSList{}
	}

	if d.Vendor == "dnspod" {
		return &TencentDNSList{}
	}

	if d.Vendor == "aliyun" {
		return &AliDNSList{}
	}

	return nil
}
