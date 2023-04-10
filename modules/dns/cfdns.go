package dns

import (
	"context"
	"domain0/models"
	lutils "domain0/utils"
	"errors"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
)

type CloudflareDNS struct {
	Id          string      `json:"id"`
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Content     string      `json:"content"`
	ProxyStatus bool        `json:"proxied"`
	TTL         int         `json:"ttl"`
	Commnet     string      `json:"comment"`
	Data        interface{} `json:"data"`
	Priority    uint16      `json:"priority"`
	Domain      models.Domain  `json:"-"`
}

type CloudflareDNSList struct {
	Success  bool            `json:"success"`
	Errors   []interface{}   `json:"errors"`
	Messages []interface{}   `json:"messages"`
	Result   []CloudflareDNS `json:"result"`
}

func (c *CloudflareDNS) Create() error {
	// extract auth info
	zoneId, apiToken, err := c.Domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Create DNS record: ", c)
	logrus.Debug("Auth with ZoneId: %s, ApiToken: %s", zoneId, apiToken)

	// create dns record
	api, err := cf.NewWithAPIToken(apiToken)
	if err != nil {
		return err
	}
	ctx := context.Background()
	record := cf.CreateDNSRecordParams{
		Type:     c.Type,
		Name:     c.Name,
		Content:  c.Content,
		TTL:      c.TTL,
		Proxied:  &c.ProxyStatus,
		Comment:  c.Commnet,
		Data:     c.Data,
		Priority: &c.Priority,
	}
	res, err := api.CreateDNSRecord(ctx, cf.ZoneIdentifier(zoneId), record)
	if err != nil {
		return err
	}

	c.Id = res.Result.ID
	return nil
}

func (c *CloudflareDNS) Get(id string) error {
	// set id
	c.Id = id

	// extract auth info
	zoneId, apiToken, err := c.Domain.ExtractAuth()
	if err != nil {
		return err
	}

	// get dns record
	api, err := cf.NewWithAPIToken(apiToken)
	if err != nil {
		return err
	}

	ctx := context.Background()
	res, err := api.GetDNSRecord(ctx, cf.ZoneIdentifier(zoneId), c.Id)
	if err != nil {
		return err
	}

	c.Type = res.Type
	c.Name = res.Name
	c.Content = res.Content
	c.ProxyStatus = lutils.IfThenPtr(res.Proxied, false)
	c.TTL = res.TTL
	c.Commnet = res.Comment
	c.Data = lutils.IfThenPtr(res.Data, "")
	c.Priority = lutils.IfThenPtr(res.Priority, uint16(0))

	// logging info
	logrus.Info("Get DNS record: ", c)
	logrus.Debug("Auth with ZoneId: %s, ApiToken: %s", zoneId, apiToken)

	return nil
}

func (c *CloudflareDNS) Delete() error {
	// extract auth info
	zoneId, apiToken, err := c.Domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Delete DNS record: ", c)
	logrus.Debug("Auth with ZoneId: %s, ApiToken: %s", zoneId, apiToken)

	// delete dns record
	api, err := cf.NewWithAPIToken(apiToken)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := api.DeleteDNSRecord(ctx, cf.ZoneIdentifier(zoneId), c.Id); err != nil {
		return err
	}
	return nil
}

func (c *CloudflareDNS) Update() error {
	// extract auth info
	zoneId, apiToken, err := c.Domain.ExtractAuth()
	if err != nil {
		return err
	}

	// logging info
	logrus.Info("Update DNS record: ", c)
	logrus.Debug("Auth with ZoneId: %s, ApiToken: %s", zoneId, apiToken)

	// update dns record
	api, err := cf.NewWithAPIToken(apiToken)
	if err != nil {
		return err
	}
	ctx := context.Background()
	record := cf.UpdateDNSRecordParams{
		ID:       c.Id,
		Type:     c.Type,
		Name:     c.Name,
		Content:  c.Content,
		TTL:      c.TTL,
		Proxied:  &c.ProxyStatus,
		Comment:  c.Commnet,
		Data:     c.Data,
		Priority: &c.Priority,
	}
	if err := api.UpdateDNSRecord(ctx, cf.ZoneIdentifier(zoneId), record); err != nil {
		return err
	}
	return nil
}

func (c *CloudflareDNSList) MultipleSelectWithIds(ids []string, r *[]interface{}) error {

	for _, dns := range c.Result {
		for _, id := range ids {
			if dns.Id == id {
				*r = append(*r, &dns)
			}
		}
	}
	if len(ids) != len(*r) {
		return errors.New("some DNS records are not found")
	}

	return nil
}

func (c *CloudflareDNSList) GetDNSList(d *models.Domain) error {
	// extract auth info
	zoneId, apiToken, err := d.ExtractAuth()
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}

	// logging info
	logrus.Info("Get DNS records of domain: %s", d.Name)
	logrus.Debug("Details: %s", d)

	// get dns record list
	api, err := cf.NewWithAPIToken(apiToken)
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}
	ctx := context.Background()
	dnsRecords, _, err := api.ListDNSRecords(ctx,
		cf.ZoneIdentifier(zoneId),
		cf.ListDNSRecordsParams{
			ResultInfo: cf.ResultInfo{PerPage: 500},
		})
	if err != nil {
		c.Errors = []interface{}{err.Error()}
		return nil
	}
	for _, dnsRecord := range dnsRecords {
		c.Result = append(c.Result, CloudflareDNS{
			Id:          dnsRecord.ID,
			Type:        dnsRecord.Type,
			Name:        dnsRecord.Name,
			Content:     dnsRecord.Content,
			ProxyStatus: lutils.IfThenPtr(dnsRecord.Proxied, false),
			TTL:         dnsRecord.TTL,
			Commnet:     dnsRecord.Comment,
			Data:        dnsRecord.Data,
			Priority:    lutils.IfThenPtr(dnsRecord.Priority, uint16(0)),
			Domain:      *d,
		})
	}
	c.Success = true
	return nil
}
