package modules

import (
	"domain0/models"
	"errors"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/sirupsen/logrus"
)

type AliDNS struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int64  `json:"ttl"`
	Commnet string `json:"comment"`
	// Data     string `json:"data"`
	Priority int64 `json:"priority"`
	domain   models.Domain
}

type AliDNSList struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []AliDNS      `json:"result"`
}

func (a *AliDNS) Create() error {
	// extract auth info
	accessKeyId, accessKeySecret, err := a.domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Create DNS record: ", a)
	logrus.Debug("Auth with AccessKeyId: %s, AccessKeySecret: %s", accessKeyId, accessKeySecret)

	// create dns record
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessKeySecret)
	if err != nil {
		return err
	}
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.DomainName = a.domain.Name
	request.RR = a.Name
	request.Type = a.Type
	request.Value = a.Content
	request.TTL = requests.NewInteger64(a.TTL)
	request.Priority = requests.NewInteger64(a.Priority)
	res, err := client.AddDomainRecord(request)
	if err != nil {
		return err
	}

	a.Id = res.RecordId
	return nil
}

func (a *AliDNS) Get(id string) error {
	// set id
	a.Id = id

	// extract auth info
	accessKeyId, accessKeySecret, err := a.domain.ExtractAuth()
	if err != nil {
		return err
	}

	// get dns record
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessKeySecret)
	if err != nil {
		return err
	}
	request := alidns.CreateDescribeDomainRecordInfoRequest()
	request.Scheme = "https"
	request.RecordId = a.Id
	res, err := client.DescribeDomainRecordInfo(request)
	if err != nil {
		return err
	}

	a.Name = res.RR
	a.Type = res.Type
	a.Content = res.Value
	a.Commnet = "interface not support yet"
	a.TTL = res.TTL
	a.Priority = res.Priority

	// logging info
	logrus.Info("Get DNS record: ", a)
	logrus.Debug("Auth with AccessKeyId: %s, AccessKeySecret: %s", accessKeyId, accessKeySecret)

	return nil
}

func (a *AliDNS) Delete() error {
	// extract auth info
	accessKeyId, accessKeySecret, err := a.domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Delete DNS record: ", a)
	logrus.Debug("Auth with AccessKeyId: %s, AccessKeySecret: %s", accessKeyId, accessKeySecret)

	// delete dns record
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessKeySecret)
	if err != nil {
		return err
	}
	request := alidns.CreateDeleteDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = a.Id
	_, err = client.DeleteDomainRecord(request)
	if err != nil {
		return err
	}

	return nil
}

func (a *AliDNS) Update() error {
	// extract auth info
	accessKeyId, accessKeySecret, err := a.domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Update DNS record: ", a)
	logrus.Debug("Auth with AccessKeyId: %s, AccessKeySecret: %s", accessKeyId, accessKeySecret)

	// update dns record
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessKeySecret)
	if err != nil {
		return err
	}

	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = a.Id
	request.RR = a.Name
	request.Type = a.Type
	request.Value = a.Content
	request.TTL = requests.NewInteger64(a.TTL)
	request.Priority = requests.NewInteger64(a.Priority)
	_, err = client.UpdateDomainRecord(request)
	if err != nil {
		return err
	}

	return nil
}

func (a *AliDNSList) MultipleSelectWithIds(ids []string, r *[]interface{}) error {
	for _, id := range ids {
		for _, dns := range a.Result {
			if dns.Id == id {
				*r = append(*r, &dns)
			}
		}
	}

	if len(*r) != len(ids) {
		return errors.New("some DNS records are not found")
	}

	return nil
}

func (c *AliDNSList) GetDNSList(d *models.Domain) error {
	// extract auth info
	accessKeyId, accessKeySecret, err := d.ExtractAuth()
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}

	// logging info
	logrus.Debug("Auth with AccessKeyId: %s, AccessKeySecret: %s", accessKeyId, accessKeySecret)

	// get dns list
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessKeySecret)
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = d.Name
	request.PageSize = requests.NewInteger(500)
	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}

	// convert to AliDNSList
	for _, record := range response.DomainRecords.Record {
		c.Result = append(c.Result, AliDNS{
			Id:      record.RecordId,
			Type:    record.Type,
			Name:    record.RR,
			Content: record.Value,
			TTL:     record.TTL,
			// Data:     record.Data,
			Priority: record.Priority,
			domain:   *d,
		})
	}

	c.Success = true
	return nil
}
