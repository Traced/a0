package temp_mailbox

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/Traced/a0/utils"
)

var (
	mailURL     = "https://mail.td"
	mailApiURL  = mailURL + "/api/api/v1/mailbox/"
	mailSuffix  = [5]string{"nqmo.com","yzm.de","qabq.com","end.tw","uuf.me"}
	mailHeaders = map[string]string{
		"authority":        "mail.td",
		"user-agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36",
		"sec-ch-ua-mobile": "?0",
	}
)

func NewMailBox() *MailBox {
	return &MailBox{}
}

type MailBox struct {
	Counter       int
	Address       string
	Token         string
	URL           string
	ApiURL        string
	QueryInterval time.Duration
}

func RandMailSuffix() string {
	return mailSuffix[rand.Intn(5)]
}

func FakerMailAddress(username string) string {
	if username == "" {
		username = faker.Username()
	}
	if 1 > strings.Index(username, "@") {
		username += RandMailSuffix()
	}
	return username
}

func (m *MailBox) GetAuthToken() string {
	resp, err := utils.GetWithHeader(nil, mailURL, mailHeaders, nil)
	if err != nil {
		log.Println("获取 auth token 错误：", err)
		return ""
	}
	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" {
			return c.Value
		}
	}
	return ""
}

func (m *MailBox) NewMailBox(username string) *MailBox {
	m.Address = FakerMailAddress(username)
	m.Token = m.GetAuthToken()
	return m
}

type (
	newMailResponse struct {
		Mailbox     string    `json:"mailbox"`
		ID          string    `json:"id"`
		From        string    `json:"from"`
		To          []string  `json:"to"`
		Subject     string    `json:"subject"`
		Date        time.Time `json:"date"`
		PosixMillis int64     `json:"posix-millis"`
		Size        int       `json:"size"`
		Seen        bool      `json:"seen"`
	}
	NewMailResponses []newMailResponse
	NewMailHandler   func(response MailBodyResponse) bool
)

func (m *MailBox) QueryNewMail() (mailList NewMailResponses) {
	mailHeaders["authorization"] = "bearer " + m.Token
	resp, err := utils.GetWithHeader(nil, mailApiURL+m.Address, mailHeaders, nil)
	if err != nil {
		log.Println("查询邮件发生错误：", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	if bytes.Index(body, []byte("expired")) > -1 {
		log.Println("查询新邮件失败，token 过期，正在获取新的 token 中")
		return m.NewMailBox(m.Address).QueryNewMail()
	}
	if err := json.Unmarshal(body, &mailList); err != nil {
		log.Println("查询新邮件失败：", err)
	}
	return
}

type MailBodyResponse struct {
	Mailbox     string    `json:"mailbox"`
	ID          string    `json:"id"`
	From        string    `json:"from"`
	To          []string  `json:"to"`
	Subject     string    `json:"subject"`
	Date        time.Time `json:"date"`
	PosixMillis int64     `json:"posix-millis"`
	Size        int       `json:"size"`
	Seen        bool      `json:"seen"`
	Body        struct {
		Text string `json:"text"`
		HTML string `json:"html"`
	} `json:"body"`
}

func (m *MailBox) GetMailBody(mailID string) (mb MailBodyResponse) {
	resp, err := utils.GetWithHeader(nil, mailApiURL+m.Address+"/"+mailID, mailHeaders, nil)
	if err != nil {
		log.Println(mailID+" 获取邮件内容失败:", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &mb); err != nil {
		log.Println(mailID+" 获取邮件内容失败:", err)
	}
	return
}

func (m *MailBox) Forever(handler NewMailHandler) {
	if m.Token == "" {
		m.NewMailBox("")
	}
	var (
		newMailList    NewMailResponses
		newMailCounter int
		interval       = m.QueryInterval
	)
	// 间隔太短，最低2秒查询一次
	if interval < time.Second<<1 {
		interval = time.Second << 1
	}
	if handler == nil {
		handler = func(r MailBodyResponse) bool {
			log.Printf("%s - %s:\n\t%s\n\t%s\n\n", r.From, r.Date, r.Subject, r.Body.HTML)
			return true
		}
	}
	log.Println("等待接收", m.Address, "的新邮件中..")
	for {
		newMailList = m.QueryNewMail()
		newMailCounter = len(newMailList) - m.Counter
		if newMailCounter > 0 {
			// 更新已有已有邮件数量
			m.Counter += newMailCounter
			log.Printf("有 %d 封新邮件", newMailCounter)
			// 倒序新邮件列表
			newMailList = newMailList[len(newMailList)-newMailCounter:]
			for end := len(newMailList) - 1; end > -1; end-- {
				info := newMailList[end]
				if !handler(m.GetMailBody(info.ID)) {
					return
				}
			}
		}
		time.Sleep(interval)
	}
}
