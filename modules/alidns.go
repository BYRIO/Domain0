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
	TTL     int    `json:"ttl"`
	Commnet string `json:"comment"`
	// Data     string `json:"data"`
	Priority int `json:"priority"`
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
	request.TTL = requests.NewInteger(a.TTL)
	request.Priority = requests.NewInteger(a.Priority)
	_, err = client.AddDomainRecord(request)
	if err != nil {
		return err
	}

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
	request.TTL = requests.NewInteger(a.TTL)
	request.Priority = requests.NewInteger(a.Priority)
	_, err = client.UpdateDomainRecord(request)
	if err != nil {
		return err
	}

	return nil
}

func (a *AliDNSList) MultipleSelectWithIds(ids []string, r *interface{}) error {
	var dnsList []AliDNS

	for _, id := range ids {
		for _, dns := range a.Result {
			if dns.Id == id {
				dnsList = append(dnsList, dns)
			}
		}
	}

	if len(dnsList) != len(ids) {
		return errors.New("some DNS records are not found")
	}

	*r = dnsList
	return nil
}

func (c *AliDNSList) GetDNSList(d *models.Domain) error {
	// extract auth info
	accessKeyId, accessKeySecret, err := d.ExtractAuth()
	if err != nil {
		c = &AliDNSList{
			Success: false,
			Errors:  []interface{}{err},
		}
		return err
	}

	// logging info
	logrus.Debug("Auth with AccessKeyId: %s, AccessKeySecret: %s", accessKeyId, accessKeySecret)

	// get dns list
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", accessKeyId, accessKeySecret)
	if err != nil {
		c = &AliDNSList{
			Success: false,
			Errors:  []interface{}{err},
		}
		return err
	}
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = d.Name
	request.PageSize = requests.NewInteger(500)
	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		c = &AliDNSList{
			Success: false,
			Errors:  []interface{}{err},
		}
		return err
	}

	// convert to AliDNSList
	var aliDNSList AliDNSList
	for _, record := range response.DomainRecords.Record {
		aliDNSList.Result = append(aliDNSList.Result, AliDNS{
			Id:      record.RecordId,
			Type:    record.Type,
			Name:    record.RR,
			Content: record.Value,
			TTL:     int(record.TTL),
			// Data:     record.Data,
			Priority: int(record.Priority),
			domain:   *d,
		})
	}

	c = &aliDNSList
	return nil
}
