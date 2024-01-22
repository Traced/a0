package yun_ti

import (
	"bytes"
	"io"
	"log"
	"regexp"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/Traced/a0/utils"
)

var (
	c         = utils.NewClient()
	apiPrefix = "http://my.yunt.pro/api"
	headers   = map[string]string{
		"content-type": "application/json",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36",
	}
	SSRRegexpMatcher = regexp.MustCompile(`(ssr://[^'"]+)`)
	defaultClient    = new(YunTi)
)

func AutoGetSSRNodes() []byte {
	nodes := defaultClient.Register().Login().GetSSRNodes()
	return bytes.Join(nodes, []byte("\n"))
}

func post(apiPath string, data io.Reader) (respBody []byte) {
	apiURL := apiPrefix + apiPath
	resp, err := utils.PostWithHeader(c, apiURL, data, headers, nil)
	if err != nil {
		log.Printf("请求发生%s错误：%s", apiURL, err)
		return
	}
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应发生%s错误：%s", apiURL, err)
		return
	}
	return
}

type YunTi struct {
	email, password, token string
	userID                 int
}

func (yt *YunTi) Register() *YunTi {
	email := utils.GetCurrentTimestampString() + "@qq.com"
	return yt.RegisterByAccount(email, email)
}

func (yt *YunTi) RegisterByAccount(email, password string) *YunTi {
	data := utils.JsonBytes(utils.StringMap{
		"email": email, "password": password, "terminal": "windows",
	})
	log.Println("注册账号：", email, "密码：", password)
	data = post("/signup", bytes.NewReader(data))
	var resp baseResponse
	_ = jsoniter.Unmarshal(data, &resp)
	log.Println(resp.Message)
	if resp.Code == 200 {
		yt.email, yt.password, yt.userID = email, password, resp.UserId
	}
	return yt
}

func (yt *YunTi) Login() *YunTi {
	return yt.LoginByAccount(yt.email, yt.password)
}

func (yt *YunTi) LoginByAccount(email, password string) *YunTi {
	data := utils.JsonBytes(utils.StringMap{
		"email": email, "password": password, "terminal": "windows",
	})
	log.Println("登录账号：", email, "密码：", password)
	data = post("/login", bytes.NewReader(data))
	var resp baseResponse
	_ = jsoniter.Unmarshal(data, &resp)
	log.Println(resp.Message)
	if resp.Code == 200 {
		yt.token = resp.Token
	}
	return yt
}

func (yt *YunTi) GetSSRNodes() (nodes [][]byte) {
	if yt.token == "" {
		log.Println("请先登录！")
		return
	}
	data := post("/users/"+strconv.Itoa(yt.userID)+"/ssrnodes", bytes.NewReader(utils.JsonBytes(utils.StringMap{
		"token": yt.token,
	})))
	var resp nodeResponse
	_ = jsoniter.Unmarshal(data, &resp)
	log.Println(resp.Message)
	if resp.Code == 200 {
		nodes = SSRRegexpMatcher.FindAll(data, -1)
	}
	return
}

