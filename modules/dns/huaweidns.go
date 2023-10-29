package dns

import (
	"domain0/models"
	"domain0/utils"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
	"github.com/sirupsen/logrus"
)

type HuaweiDNS struct {
	Id      string   `json:"id"`
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Content []string `json:"content"`
	TTL     int      `json:"ttl"`
	Commnet string   `json:"comment"`
	//Data     interface{}   `json:"data"`
	//Priority uint16        `json:"priority"`
	Domain models.Domain `json:"-"`
}

type HuaweiDNSList struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []HuaweiDNS   `json:"result"`
}

func (h *HuaweiDNS) GetZoneId(domainName string) (string, error) {
	ak, sk, err := h.Domain.ExtractAuth()
	if err != nil {
		return "", err
	}
	// auth
	auth := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())

	request := &model.ListPublicZonesRequest{}
	nameRequest := domainName
	request.Name = &nameRequest
	response, err := client.ListPublicZones(request)
	if err == nil {
		for _, zone := range *response.Zones {
			if *zone.Name == domainName+"." {
				return *zone.Id, nil
			}
		}
	}
	return "", nil
}

func (h *HuaweiDNS) Create() error {
	// extract auth info
	// ak->accessid, sk->accesskey
	ak, sk, err := h.Domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Create DNS record: ", h)
	logrus.Debug("Auth with Secret_Id: %s, Secret_Key: %s", ak, sk)

	// auth
	auth := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())

	// create dns record
	request := &model.CreateRecordSetRequest{}
	ttlRecord := int32(h.TTL)
	//mxRecord := strconv.Itoa(int(h.Priority)) + " " + h.Content
	zoneID, err := h.GetZoneId(h.Domain.Name)
	request.ZoneId = zoneID
	request.Body = &model.CreateRecordSetRequestBody{
		Name:        h.Name + "." + h.Domain.Name + ".",
		Type:        h.Type,
		Ttl:         utils.IfThen(h.TTL == 0, nil, &ttlRecord),
		Records:     h.Content,
		Description: &h.Commnet,
	}
	res, err := client.CreateRecordSet(request)
	if err != nil {
		return err
	}
	h.Id = *res.Id
	return nil
}

func (h *HuaweiDNS) Get(id string) error {
	// set id
	h.Id = id

	// extract auth info
	// ak->accessid, sk->accesskey
	ak, sk, err := h.Domain.ExtractAuth()
	if err != nil {
		return err
	}

	// auth
	auth := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())

	// get dns record
	request := &model.ShowRecordSetRequest{}
	zoneID, err := h.GetZoneId(h.Domain.Name)
	request.ZoneId = zoneID
	request.RecordsetId = h.Id
	res, err := client.ShowRecordSet(request)
	if err != nil {
		return err
	}
	h.Name = *res.Name
	h.Type = *res.Type
	h.Content = *res.Records
	h.TTL = int(*res.Ttl)
	h.Commnet = *res.Description

	return nil
}

func (h *HuaweiDNS) Delete() error {
	// extract auth info
	// ak->accessid, sk->accesskey
	ak, sk, err := h.Domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Delete DNS record: ", h)
	logrus.Debug("Auth with Secret_Id: %s, Secret_Key: %s", ak, sk)

	// auth
	auth := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())

	// delete dns record
	request := &model.DeleteRecordSetRequest{}
	zoneID, err := h.GetZoneId(h.Domain.Name)
	request.ZoneId = zoneID
	request.RecordsetId = h.Id
	if _, err := client.DeleteRecordSet(request); err != nil {
		return err
	}

	return nil
}

func (h *HuaweiDNS) Update() error {
	// extract auth info
	ak, sk, err := h.Domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Update DNS record: ", h)
	logrus.Debug("Auth with Secret_Id: %s, Secret_Key: %s", ak, sk)

	// auth
	auth := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())

	// update dns record
	request := &model.UpdateRecordSetRequest{}
	request.ZoneId = h.Domain.Name
	request.RecordsetId = h.Id
	ttlRecord := int32(h.TTL)
	ttl300 := int32(300)
	//mxRecord := strconv.Itoa(int(h.Priority)) + " " + h.Content
	zoneID, err := h.GetZoneId(h.Domain.Name)
	request.ZoneId = zoneID
	request.Body = &model.UpdateRecordSetReq{
		Name:        h.Name + "." + h.Domain.Name + ".",
		Type:        h.Type,
		Ttl:         utils.IfThen(h.TTL == 0, &ttl300, &ttlRecord),
		Records:     &h.Content,
		Description: &h.Commnet,
	}
	_, err = client.UpdateRecordSet(request)
	if err != nil {
		return err
	}
	return nil
}

func (h *HuaweiDNSList) MultipleSelectWithIds(ids []string, r *[]interface{}) error {
	for _, id := range ids {
		for _, record := range h.Result {
			if record.Id == id {
				*r = append(*r, &record)
			}
		}
	}
	if len(*r) != len(ids) {
		return errors.New("some dns record not found")
	}

	return nil
}

func (h *HuaweiDNSList) GetDNSList(d *models.Domain) error {
	// extract auth info
	ak, sk, err := d.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Get DNS record list: ", d)
	logrus.Debug("Auth with Secret_Id: %s, Secret_Key: %s", ak, sk)

	// auth
	auth := basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()

	client := dns.NewDnsClient(
		dns.DnsClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())

	// get dns record list
	request := &model.ListRecordSetsRequest{}
	nameRequest := d.Name
	request.Name = &nameRequest
	response, err := client.ListRecordSets(request)
	if err != nil {
		h.Errors = []interface{}{err.Error()}
		return nil
	}
	for _, record := range *response.Recordsets {
		h.Result = append(h.Result, HuaweiDNS{
			Id:      *record.Id,
			Type:    *record.Type,
			Name:    *record.Name,
			Content: *record.Records,
			TTL:     int(*record.Ttl),
			Domain:  *d,
			Commnet: *record.Description,
		})
	}
	h.Success = true
	return nil
}
