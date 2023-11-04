package dns

import (
	"crypto/md5"
	"domain0/models"
	"domain0/utils"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
	"github.com/sirupsen/logrus"
)

// type HuaweiDNSCustom struct{}

// TODO: It would be best to use reflection to reconstruct this logic
type sourceRecord struct {
	list   *model.ListRecordSetsWithTags
	record *model.ShowRecordSetResponse
}

func (s *sourceRecord) getRecords() *[]string {
	if s.list != nil {
		return s.list.Records
	}
	if s.record != nil {
		return s.record.Records
	}
	return nil
}

func (s *sourceRecord) getId() *string {
	if s.list != nil {
		return s.list.Id
	}
	if s.record != nil {
		return s.record.Id
	}
	return nil
}

func (s *sourceRecord) getName() *string {
	if s.list != nil {
		return s.list.Name
	}
	if s.record != nil {
		return s.record.Name
	}
	return nil
}
func (s *sourceRecord) getType() *string {
	if s.list != nil {
		return s.list.Type
	}
	if s.record != nil {
		return s.record.Type
	}
	return nil
}
func (s *sourceRecord) getTtl() *int32 {
	if s.list != nil {
		return s.list.Ttl
	}
	if s.record != nil {
		return s.record.Ttl
	}
	return nil
}
func (s *sourceRecord) getDescription() *string {
	if s.list != nil {
		return s.list.Description
	}
	if s.record != nil {
		return s.record.Description
	}
	return nil
}

// sourceRecord: end

type HuaweiDNS struct {
	Id      string  `json:"id"`
	Type    string  `json:"type"`
	Name    string  `json:"name"`
	Content string  `json:"content"`
	TTL     int     `json:"ttl"`
	Commnet *string `json:"comment"`
	//Data     interface{}   `json:"data"`
	Priority uint16 `json:"priority"`
	// Custom HuaweiDNSCustom `json:"custom"`
	Domain models.Domain `json:"-"`

	sourceRecord *sourceRecord
	client       *dns.DnsClient
	zoneId       *string
}

type HuaweiDNSList struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []HuaweiDNS   `json:"result"`
}

func hashRecord(records *[]string, index int) string {
	record := (*records)[index]
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(record))
	md5Value := md5Ctx.Sum(nil)
	orderNumber := 0
	for _, item := range (*records)[:index] {
		if item == record {
			orderNumber += 1
		}
	}
	return hex.EncodeToString(append(md5Value[:6], byte(orderNumber)))
}

// HW id:  HW<id from Huawei SDK>$$<hash of the record content>
func encodeId(baseId, hash string) string {
	return fmt.Sprintf("HW%s@%s", baseId, hash)
}

func decodeId(id string) (string, string, error) {
	if len(id) <= 2 || id[:2] != "HW" {
		return "", "", errors.New("'id' is an invalid value")
	}
	id_slice := strings.Split(id[2:], "@")
	if len(id_slice) != 2 {
		return "", "", errors.New("'id' does not contain a valid id")
	}
	return id_slice[0], id_slice[1], nil
}

func (h *HuaweiDNS) formatName() string {
	name := h.Name
	if !strings.Contains(name, h.Domain.Name) {
		if name != "" && name[len(name)-1] != '.' && name != "@" {
			name += "."
		}
		name += h.Domain.Name + "."
	}
	h.Name = name

	return name
}

func (h *HuaweiDNS) getSourceRecord() (*sourceRecord, error) {
	if h.sourceRecord != nil {
		return h.sourceRecord, nil
	}

	name := h.formatName()
	client, err := h.getClient()
	if err != nil {
		return nil, err
	}

	sdkRecordId, _, err := decodeId(h.Id)

	if err != nil {
		// search mode like
		requestGet := &model.ListRecordSetsRequest{}

		requestGet.Name = &name
		requestGet.Type = &h.Type

		searchMode := "equal"
		requestGet.SearchMode = &searchMode

		sourceRes, err := client.ListRecordSets(requestGet)
		if err != nil {
			return nil, err
		}

		for _, record := range *sourceRes.Recordsets {
			if *record.Name == name && *record.Type == h.Type {
				h.sourceRecord = &sourceRecord{list: &record}
				return h.sourceRecord, nil
			}
		}
	} else {
		// get record set
		request := &model.ShowRecordSetRequest{}
		zoneId, err := h.getZoneId()
		if err != nil {
			return nil, err
		}
		request.ZoneId = *zoneId
		request.RecordsetId = sdkRecordId
		record, err := client.ShowRecordSet(request)
		if err != nil {
			return nil, err
		}
		h.sourceRecord = &sourceRecord{record: record}
		return h.sourceRecord, nil
	}
	return nil, nil
}

func (h *HuaweiDNS) getRecordIndex() (int, error) {

	_, recordHash, err := decodeId(h.Id)
	if err != nil {
		return -1, err
	}

	sourceRecord, err := h.getSourceRecord()
	if err != nil {
		return -1, err
	}
	if sourceRecord == nil {
		return -1, errors.New("DNS records is not found")
	}

	records := sourceRecord.getRecords()

	recordIndex := -1
	for index := 0; index < len(*records); index++ {
		if hashRecord(records, index) == recordHash {
			recordIndex = index
			break
		}
	}
	if recordIndex == -1 {
		return -1, errors.New("DNS record is not found")
	}
	return recordIndex, nil
}

func (h *HuaweiDNS) getClient() (*dns.DnsClient, error) {
	if h.client != nil {
		return h.client, nil
	}
	// extract auth info
	// ak->accessid, sk->accesskey
	ak, sk, err := h.Domain.ExtractAuth()
	if err != nil {
		return nil, err
	}

	// logging info
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

	h.client = client

	return client, nil
}

func (h *HuaweiDNS) getZoneId() (*string, error) {
	if h.zoneId != nil {
		return h.zoneId, nil
	}

	client, err := h.getClient()
	if err != nil {
		return nil, err
	}

	request := &model.ListPublicZonesRequest{}
	request.Name = &h.Domain.Name
	response, err := client.ListPublicZones(request)
	if err == nil {
		for _, zone := range *response.Zones {
			if *zone.Name == h.Domain.Name+"." {
				h.zoneId = zone.Id
				return h.zoneId, nil
			}
		}
	}
	return nil, errors.New("invalid domain name to get ZoneId")
}

func (h *HuaweiDNS) setRecord(record string) {
	switch h.Type {
	case "MX":
		content_split := strings.Split(record, " ")
		if len(content_split) == 2 {
			h.Content = content_split[1]
			priority, err := strconv.Atoi(content_split[0])
			if err == nil {
				h.Priority = uint16(priority)
			}
		} else {
			h.Content = record
			h.Priority = 0
		}
	case "TXT":
		h.Content = strings.ReplaceAll(strings.ReplaceAll(record, "\\\"", "\""), "\\\\", "\\")
		if len(h.Content) >= 2 {
			h.Content = h.Content[1:]
			h.Content = h.Content[:len(h.Content)-1]
		}
		h.Priority = 0
	default:
		h.Content = record
		h.Priority = 0
	}
}

func (h *HuaweiDNS) getRecord() string {
	switch h.Type {
	case "MX":
		return strconv.Itoa(int(h.Priority)) + " " + h.Content
	case "TXT":
		return "\"" + strings.ReplaceAll(strings.ReplaceAll(h.Content, "\\", "\\\\"), "\"", "\\\"") + "\""
	default:
		return h.Content
	}
}

func (h *HuaweiDNS) Create() error {
	logrus.Info("Create DNS record: ", h)
	// create dns record
	name := h.formatName()
	ttlRecord := int32(h.TTL)
	zoneID, err := h.getZoneId()
	if err != nil {
		return err
	}
	client, err := h.getClient()
	if err != nil {
		return err
	}

	// First load all parses and find the parsing records that need to be merged.
	sourceRecord, err := h.getSourceRecord()
	if err != nil {
		return err
	}

	if sourceRecord != nil {
		// insert records into existing record
		request := &model.UpdateRecordSetRequest{}
		request.ZoneId = *zoneID
		request.RecordsetId = *sourceRecord.getId()
		ttl300 := int32(300)
		records := append(*sourceRecord.getRecords(), h.getRecord())
		request.Body = &model.UpdateRecordSetReq{
			Name:        name,
			Type:        h.Type,
			Ttl:         utils.IfThen(h.TTL == 0, &ttl300, &ttlRecord),
			Records:     &records,
			Description: h.Commnet,
		}
		res, err := client.UpdateRecordSet(request)
		if err != nil {
			return err
		}
		h.Id = encodeId(*res.Id, hashRecord(res.Records, len(records)-1))
		h.TTL = int(*res.Ttl)
		h.Name = *res.Name
		h.Commnet = res.Description
	} else {
		// create new records
		request := &model.CreateRecordSetRequest{}
		request.ZoneId = *zoneID
		records := []string{h.getRecord()}
		request.Body = &model.CreateRecordSetRequestBody{
			Name:        name,
			Type:        h.Type,
			Ttl:         utils.IfThen(h.TTL == 0, nil, &ttlRecord),
			Records:     records,
			Description: h.Commnet,
		}
		res, err := client.CreateRecordSet(request)
		if err != nil {
			return err
		}
		h.Id = encodeId(*res.Id, hashRecord(res.Records, 0))
		h.TTL = int(*res.Ttl)
		h.Name = *res.Name
		h.Commnet = res.Description
	}

	return nil
}

func (h *HuaweiDNS) Get(id string) error {
	// set id
	h.Id = id
	logrus.Info("Get DNS record: ", h)

	_, recordHash, err := decodeId(h.Id)
	if err != nil {
		return err
	}
	// get dns record
	sourceRecord, err := h.getSourceRecord()
	if err != nil {
		return err
	}
	if sourceRecord == nil {
		return errors.New("DNS records is not found")
	}
	records := sourceRecord.getRecords()
	for index, record := range *records {
		if hashRecord(records, index) == recordHash {
			h.Name = *sourceRecord.getName()
			h.Type = *sourceRecord.getType()
			h.setRecord(record)
			h.TTL = int(*sourceRecord.getTtl())
			h.Commnet = sourceRecord.getDescription()
			return nil
		}
	}
	return errors.New("DNS record is not found")
}

func (h *HuaweiDNS) Delete() error {
	logrus.Info("Delete DNS record: ", h)
	recordIndex, err := h.getRecordIndex()
	if err != nil {
		return err
	}
	// after getRecordIndex, this will never return error
	client, _ := h.getClient()
	zoneID, _ := h.getZoneId()
	sdkRecordId, _, _ := decodeId(h.Id)
	sourceRecord, _ := h.getSourceRecord()
	records := sourceRecord.getRecords()

	if len(*records) == 1 {
		// delete dns record
		request := &model.DeleteRecordSetRequest{}
		request.ZoneId = *zoneID
		request.RecordsetId = sdkRecordId
		if _, err := client.DeleteRecordSet(request); err != nil {
			return err
		}
	} else {
		// update dns record
		request := &model.UpdateRecordSetRequest{}
		request.ZoneId = *zoneID
		request.RecordsetId = *sourceRecord.getId()
		records := append((*records)[:recordIndex], (*records)[recordIndex+1:]...)
		request.Body = &model.UpdateRecordSetReq{
			Name:        *sourceRecord.getName(),
			Type:        *sourceRecord.getType(),
			Ttl:         sourceRecord.getTtl(),
			Records:     &records,
			Description: sourceRecord.getDescription(),
		}
		_, err = client.UpdateRecordSet(request)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *HuaweiDNS) Update() error {
	logrus.Info("Update DNS record: ", h)
	name := h.formatName()
	recordIndex, err := h.getRecordIndex()
	if err != nil {
		return err
	}
	// after getRecordIndex, this will never return error
	client, _ := h.getClient()
	zoneID, _ := h.getZoneId()
	sdkRecordId, _, _ := decodeId(h.Id)
	sourceRecord, _ := h.getSourceRecord()
	records := sourceRecord.getRecords()

	// if change name or type
	// you should create new record and remove the content from the source records
	if name != *sourceRecord.getName() || h.Type != *sourceRecord.getType() {
		removeDns := &HuaweiDNS{
			// create will edit the id, so create this first
			Id:           h.Id,
			Type:         *sourceRecord.getType(),
			Name:         *sourceRecord.getName(),
			Domain:       h.Domain,
			sourceRecord: h.sourceRecord,
			client:       h.client,
			zoneId:       h.zoneId,
		}
		// remove id and sourceRecord, let Create to research the source record
		h.Id = ""
		h.sourceRecord = nil
		err := h.Create()
		if err != nil {
			return err
		}
		// remove then
		err = removeDns.Delete()
		if err != nil {
			return err
		}
	} else {
		records := append(append((*records)[:recordIndex], h.getRecord()), (*records)[recordIndex+1:]...)

		// update dns record
		request := &model.UpdateRecordSetRequest{}
		request.ZoneId = h.Domain.Name
		request.RecordsetId = sdkRecordId
		ttlRecord := int32(h.TTL)
		request.ZoneId = *zoneID
		request.Body = &model.UpdateRecordSetReq{
			Name:        *sourceRecord.getName(),
			Type:        *sourceRecord.getType(),
			Ttl:         utils.IfThen(h.TTL == 0, sourceRecord.getTtl(), &ttlRecord),
			Records:     &records,
			Description: h.Commnet,
		}
		res, err := client.UpdateRecordSet(request)
		if err != nil {
			return err
		}
		h.Id = encodeId(*res.Id, hashRecord(res.Records, recordIndex))
		h.TTL = int(*res.Ttl)
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
	request.Name = &d.Name
	response, err := client.ListRecordSets(request)
	if err != nil {
		h.Errors = []interface{}{err.Error()}
		return nil
	}
	for _, record := range *response.Recordsets {
		for index, recordItem := range *record.Records {
			recordHash := hashRecord(record.Records, index)
			// TODO: need to add "Line type"
			dnsItem := HuaweiDNS{
				Id:      encodeId(*record.Id, recordHash),
				Type:    *record.Type,
				Name:    *record.Name,
				TTL:     int(*record.Ttl),
				Domain:  *d,
				Commnet: record.Description,
			}
			dnsItem.setRecord(recordItem)
			h.Result = append(h.Result, dnsItem)
		}
	}
	h.Success = true
	return nil
}
