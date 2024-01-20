package models

const (
	MsgTypePost             = "post"
	ParagraphContentTagText = "text"
)

type Message struct {
	MsgType string  `json:"msg_type"`
	Content Content `json:"content"`
}

type Content struct {
	Post *Post `json:"post,omitempty"`
}

type Post struct {
	ZhCn *PostContent `json:"zh_cn,omitempty"`
	EnUs *PostContent `json:"en_us,omitempty"`
}

type PostContent struct {
	Title   string      `json:"title"`
	Content []Paragraph `json:"content"`
}

type Paragraph []ParagraphContent

type ParagraphContent struct {
	Tag  string `json:"tag"`
	Text string `json:"text,omitempty"`
	Href string `json:"href,omitempty"`
}
