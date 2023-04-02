package modules

import (
	"domain0/models"
	"domain0/utils"
	"errors"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

var dnsProfile = profile.NewClientProfile()

type TencentDNSCustom struct {
	RecordLine string `json:"record_line"`
	Enable     string `json:"enable"`
}

type TencentDNS struct {
	Id      uint64  `json:"id"`
	Type    string  `json:"type"`
	Name    string  `json:"name"`
	Content string  `json:"content"`
	TTL     uint64  `json:"ttl"`
	Commnet *string `json:"comment"`
	// Data     string `json:"data"`
	Priority uint64            `json:"priority"`
	Custom   *TencentDNSCustom `json:"custom"`
	domain   models.Domain
}

type TencentDNSList struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []TencentDNS  `json:"result"`
}

func (t *TencentDNS) Create() error {
	// extract auth info
	secretId, secretKey, err := t.domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Create DNS record: ", t)
	logrus.Debug("Auth with Secret_Id: %s, Secret_Key: %s", secretId, secretKey)

	// create dns record
	client, err := dnspod.NewClient(common.NewCredential(secretId, secretKey), "ap-guangzhou", dnsProfile)
	if err != nil {
		return err
	}
	request := dnspod.NewCreateRecordRequest()
	request.Domain = &t.domain.Name
	request.SubDomain = &t.Name
	request.RecordType = &t.Type
	request.RecordLine = common.StringPtr(utils.IfThen(t.Custom.RecordLine == "", "默认", t.Custom.RecordLine))
	request.Value = &t.Content
	request.TTL = utils.IfThen(t.TTL == 0, nil, &t.TTL)
	request.MX = utils.IfThen(t.Priority == 0, nil, &t.Priority)
	request.Status = common.StringPtr(utils.IfThen(t.Custom.Enable == "", "enable", t.Custom.Enable))

	res, err := client.CreateRecord(request)
	if err != nil {
		return err
	}

	t.Id = *res.Response.RecordId
	return nil
}

func (t *TencentDNS) Delete() error {
	// extract auth info
	secretId, secretKey, err := t.domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Delete DNS record: ", t)
	logrus.Debug("Auth with Secret_Id: %s, Secret_Key: %s", secretId, secretKey)

	// delete dns record
	client, err := dnspod.NewClient(common.NewCredential(secretId, secretKey), "ap-guangzhou", dnsProfile)
	if err != nil {
		return err
	}
	request := dnspod.NewDeleteRecordRequest()
	request.Domain = &t.domain.Name
	request.RecordId = &t.Id

	if _, err := client.DeleteRecord(request); err != nil {
		return err
	}

	return nil
}

func (t *TencentDNS) Update() error {
	// extract auth info
	secretId, secretKey, err := t.domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Update DNS record: ", t)
	logrus.Debug("Auth with Secret_Id: %s, Secret_Key: %s", secretId, secretKey)

	// update dns record
	client, err := dnspod.NewClient(common.NewCredential(secretId, secretKey), "ap-guangzhou", dnsProfile)
	if err != nil {
		return err
	}
	request := dnspod.NewModifyRecordRequest()
	request.Domain = &t.domain.Name
	request.RecordType = &t.Type
	request.RecordId = &t.Id
	request.SubDomain = &t.Name
	request.RecordLine = common.StringPtr(utils.IfThen(t.Custom.RecordLine == "", "默认", t.Custom.RecordLine))
	request.Value = &t.Content
	request.TTL = utils.IfThen(t.TTL == 0, nil, &t.TTL)
	request.MX = utils.IfThen(t.Priority == 0, nil, &t.Priority)
	request.Status = common.StringPtr(utils.IfThen(t.Custom.Enable == "", "enable", t.Custom.Enable))

	if _, err := client.ModifyRecord(request); err != nil {
		return err
	}

	reamrkRequest := dnspod.NewModifyRecordRemarkRequest()

	if t.Commnet != nil {
		reamrkRequest.Domain = &t.domain.Name
		reamrkRequest.RecordId = &t.Id
		reamrkRequest.Remark = t.Commnet

		client.ModifyRecordRemark(reamrkRequest) // 是否成功就不怎么重要了
	}

	return nil
}

func (t *TencentDNSList) MultipleSelectWithIds(ids []string, r *[]interface{}) error {

	for _, id := range ids {
		for _, dns := range t.Result {
			if strconv.Itoa(int(dns.Id)) == id {
				*r = append(*r, &dns)
			}
		}
	}

	if len(*r) != len(ids) {
		return errors.New("some dns record not found")
	}

	return nil
}

func (c *TencentDNSList) GetDNSList(d *models.Domain) error {
	// extract auth info
	secretId, secretKey, err := d.ExtractAuth()
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}

	// logging info
	logrus.Info("Get DNS record list: ", d)
	logrus.Debug("Auth with Secret_Id: %s, Secret_Key: %s", secretId, secretKey)

	// get dns record list
	api, err := dnspod.NewClient(common.NewCredential(secretId, secretKey), "ap-guangzhou", dnsProfile)
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}

	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = &d.Name
	request.Limit = common.Uint64Ptr(500)
	response, err := api.DescribeRecordList(request)
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}

	for _, record := range response.Response.RecordList {
		c.Result = append(c.Result, TencentDNS{
			Id:       *record.RecordId,
			Type:     *record.Type,
			Name:     *record.Name,
			Content:  *record.Value,
			TTL:      *record.TTL,
			Commnet:  record.Remark,
			Priority: utils.IfThen(record.MX == nil, 0, *record.MX),
			Custom:   &TencentDNSCustom{Enable: *record.Status, RecordLine: *record.Line},
			domain:   *d,
		})
	}

	c.Success = true
	return nil
}
