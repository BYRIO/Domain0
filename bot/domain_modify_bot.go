package bot

import (
	"bytes"
	"domain0/bot/models"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

const (
	notifyDomainModifyFmt = "操作人: %s\n域名: %s\n" +
		"操作请求url: %s\n时间: %s\nResult: %s"
	postTitle           = "域名修改通知"
	modifyResultSuccess = "Success"
	modifyResultFail    = "Fail"
)

type DomainModifyRecord struct {
	UserName      string
	Domain        string
	OperationUrl  string
	OperationTime time.Time
	Result        bool
}

var client *http.Client

func init() {
	// do not use default http client!
	// for detail: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	client = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

}

func NotifyDomainChange(domainNotifyRecords []DomainModifyRecord) {
	var paragraphs []string
	for _, record := range domainNotifyRecords {
		notifyTxt := constructNotifyTxt(record)
		paragraphs = append(paragraphs, notifyTxt)
	}

	request, err := constructHttpRequest(postTitle, paragraphs)
	response, err := client.Do(request)
	if err != nil {
		logrus.Errorf("notify Feishu bot error:%v", err)
		return
	}
	if response.StatusCode/100 != 2 {
		logrus.Errorf("Feishu response error:%v", response)
		return
	}
}

func constructNotifyTxt(record DomainModifyRecord) string {
	var result string
	if record.Result {
		result = modifyResultSuccess
	} else {
		result = modifyResultFail
	}
	return fmt.Sprintf(notifyDomainModifyFmt, record.UserName, record.Domain,
		record.OperationUrl, record.OperationTime.Format("2006-01-02T15:04:05 -070000"),
		result)
}

func constructHttpRequest(title string, paragraphs []string) (*http.Request, error) {
	requestBodyForRichText := constructRequestBodyForRichText(title, paragraphs)
	requestBodyJson, err := json.Marshal(requestBodyForRichText)
	if err != nil {
		return nil, err
	}
	url := "https://open.feishu.cn/open-apis/bot/v2/hook/f6a7fef9-dfe1-4311-b874-2a059c73ac28"
	return http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBodyJson))
}

func constructRequestBodyForRichText(title string, paragraphContents []string) models.Message {
	var paragraphs []models.Paragraph
	for _, content := range paragraphContents {
		paragraphContent := models.ParagraphContent{Tag: models.ParagraphContentTagText, Text: content}
		paragraphs = append(paragraphs, models.Paragraph{paragraphContent})
	}
	post := models.Post{ZhCn: &models.PostContent{Title: title, Content: paragraphs}}
	res := models.Message{
		MsgType: models.MsgTypePost,
		Content: models.Content{Post: &post},
	}
	return res
}
