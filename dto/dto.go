package dto

// OptionPair 选项对（支持富媒体：文本/图片/视频）
type OptionPair struct {
	Label   string `json:"label"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
}

// RichOption 富媒体选项请求结构
type RichOption struct {
	Label   string `json:"label"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
}
