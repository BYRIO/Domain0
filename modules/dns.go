package modules

import (
	"domain0/models"
	. "domain0/modules/dns"
)

type DnsObj interface {
	Create() error
	Get(id string) error
	Update() error
	Delete() error
}

type DnsObjList interface {
	GetDNSList(d *models.Domain) error
	MultipleSelectWithIds(ids []string, r *[]interface{}) error
}

type DnsChangeStruct struct {
	Dns    DnsObj        `json:"dns"`
	Domain models.Domain `json:"domain"`
}

func DnsObjGen(d *models.Domain) DnsObj {
	if d.Vendor == "cloudflare" {
		return &CloudflareDNS{Domain: *d}
	}

	if d.Vendor == "dnspod" {
		return &TencentDNS{Domain: *d}
	}

	if d.Vendor == "aliyun" {
		return &AliDNS{Domain: *d}
	}

	if d.Vendor == "huawei" {
		return &HuaweiDNS{Domain: *d}
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

	if d.Vendor == "huawei" {
		return &HuaweiDNSList{}
	}
	return nil
}

func (dcs *DnsChangeStruct) DnsChangeRestore() error {
	if dcs.Domain.Vendor == "cloudflare" {
		dcs.Dns.(*CloudflareDNS).Domain = dcs.Domain
	} else if dcs.Domain.Vendor == "dnspod" {
		dcs.Dns.(*TencentDNS).Domain = dcs.Domain
	} else if dcs.Domain.Vendor == "aliyun" {
		dcs.Dns.(*AliDNS).Domain = dcs.Domain
	} else if dcs.Domain.Vendor == "huawei" {
		dcs.Dns.(*HuaweiDNS).Domain = dcs.Domain
	}
	return nil
}
