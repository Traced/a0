package astar_vpn

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"crypto/rc4"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Traced/a0/app/astar_vpn/protected/xray"
	"github.com/Traced/a0/app/astar_vpn/utils"
)

var (
	secret = "H+t2uEOGt3NsiLKOekLh7fX8K0mpbrx6VoauN2mXW/h6nV+RcPngxC98dET7"
	key    = []byte("519782632662441055c2c0990cec53b5")
	params = map[string]string{
		"strP":         "lldamljbjndnopbjooolkdnnfidlcmjk",
		"nonlinestate": "1",
		"version":      "PC109",
		"clientUUID":   "pc-61786c73-851f-47bd-8f2a-597354b0edcd",
		"apiSource":    "pc",
	}
	headers = map[string]string{
		"Connection":      "keep-alive",
		"Accept":          "application/json, text/plain, */*",
		"User-Agent":      "Electron Desktop UnUnPC macOS x64",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "zh-CN",
		"Content-Type":    "application/x-www-form-urlencoded",
	}
	mailbox         = NewMailBox()
	activeLinkMatch = regexp.MustCompile("(http.+) ")
	useProxy        bool
)

func NewAStarVPNClient(account, password string) *AStarVPNClient {
	params["strlognid"] = account
	headers["Host"] = GetApiHost()
	return &AStarVPNClient{
		Account:  account,
		Password: password,
		ApiHost:  headers["Host"],
		ApiURL:   GetApiURL(),
	}
}

func GetApiURL() string {
	return "https://" + GetApiHost() + "/astarnew/NewVPN"
}

func GetApiHost() string {
	subdomain := time.Now().Format("01023")
	return "w" + subdomain + "s.astarvpn.center"
}

type AStarVPNClient struct {
	Account, Password string
	ApiHost, ApiURL   string
}

func (a *AStarVPNClient) ActiveAccount(account string) *AStarVPNClient {
	if account == "" {
		account = a.Account
	}
	mailbox.NewMailBox(account).Forever(func(r MailBodyResponse) bool {
		if !strings.Contains(r.From, "astarvpn") {
			return true
		}
		var (
			match = activeLinkMatch.FindStringSubmatch(r.Body.Text)
			keep  = 1 > len(match)
		)
		if !keep {
			_, _ = utils.Get(match[0])
			log.Println("账号激活成功：", a.Account)
		}
		return keep
	})
	return a
}

func (a *AStarVPNClient) RandomSetAccount() *AStarVPNClient {
	a.Account = FakerMailAddress("")
	return a
}

func (a *AStarVPNClient) Register() *AStarVPNClient {
	if a.Account == "" {
		a.RandomSetAccount()
	}
	log.Println("使用邮箱注册：", a.Account)
	params["strlognid"] = a.Account
	params["strpassword"] = a.Password
	params["strvcode"] = "123456"
	data := utils.StringMapToBufferReader(params)
	apiURL := a.ApiURL[:len(a.ApiURL)-7] + "/user/register"
	useProxy = true
	resp, err := utils.PostWithHeader(apiURL, data, headers)
	useProxy = false
	if err != nil {
		log.Println("注册失败：", err)
		return a
	}
	log.Println("读取注册响应：")
	r, _ := io.ReadAll(resp.Body)
	if 1 > bytes.Index(r, []byte("successful")) {
		log.Println("注册失败：", string(r))
		return a
	}
	log.Println("注册成功：", a.Account)
	return a.ActiveAccount(a.Account)
}

type (
	ProxyListResponse struct {
		StrText    string `json:"strText"`
		NCode      int    `json:"nCode"`
		JSONObject struct {
			D EndPoints `json:"d"`
			I string    `json:"i"`
		} `json:"jsonObject"`
	}
	EndPoint struct {
		VipType  int    `json:"pr"`
		URL      string `json:"c"`
		Name     string `json:"n"`
		ID       int    `json:"i"`
		StrToken string
	}
	EndPoints     []EndPoint
	ProxyNode     xray.BaseNodeConfig
	ProxyResponse struct {
		StrText string     `json:"str_text"`
		Code    int        `json:"nCode"`
		Data    *ProxyNode `json:"jsonObject"`
		Name    string
	}
)

func (a *AStarVPNClient) GetProxyEndPoints() (eps EndPoints) {
	var (
		url  = a.ApiURL + "/getProxyList"
		data = utils.StringMapToBufferReader(params)
	)
	rsp, err := utils.PostWithHeader(url, data, headers)
	if err != nil {
		log.Fatalln("请求代理列表错误：", err)
		return
	}
	var (
		body, _          = io.ReadAll(rsp.Body)
		bodyString, resp = DecodeResponse(body)
		plr              ProxyListResponse
	)
	if err = json.Unmarshal([]byte(bodyString), &plr); err != nil {
		log.Println("解析代理列表失败：", err)
		return
	}
	for _, item := range plr.JSONObject.D {
		item.StrToken = resp.S
		eps = append(eps, item)
	}
	return
}

func (a *AStarVPNClient) GetXrayConfigByProxyNode(pn ProxyNode) xray.Config {
	return xray.GetConfigByBaseNodeConfig(xray.BaseNodeConfig(pn))
}

func (a *AStarVPNClient) ProxyList() (pl []ProxyNode) {
	log.Println("正在获取可用代理列表")
	var (
		eps = a.GetProxyEndPoints()
		epl = len(eps)
		pc  = make(chan ProxyResponse, epl)
	)
	pl = make([]ProxyNode, 0, epl)
	log.Printf("共获取到%d个测试点", epl)
	// 并发获取
	for _, e := range eps {
		go func(ed EndPoint) {
			pc <- a.GetProxy(ed)
		}(e)
	}
	for range eps {
		p := <-pc
		if p.Data != nil {
			pl = append(pl, *p.Data)
		}
	}
	log.Printf("获取到 %d 个代理路线", len(pl))
	return
}

func (a *AStarVPNClient) GetProxy(e EndPoint) (pr ProxyResponse) {
	var (
		data = utils.StringMapToBufferReader(map[string]string{
			"strP":         "lldamljbjndnopbjooolkdnnfidlcmjk",
			"strlognid":    a.Account,
			"nonlinestate": "1",
			"version":      "PC109",
			"strtoken":     e.StrToken,
			"lid":          strconv.Itoa(e.ID),
			"apiSource":    "pc",
		})
		resp, _     = utils.PostWithHeader(a.ApiURL+"/getProxy", data, headers)
		encBody, _  = io.ReadAll(resp.Body)
		dataBody, _ = DecodeResponse(encBody)
		err         = json.Unmarshal([]byte(dataBody), &pr)
	)
	pr.Name = e.Name
	if err != nil {
		log.Println("获取代理信息返回的响应体解析失败：", e.Name, err, dataBody)
	}
	if pr.Data != nil {
		pr.Data.Name = e.Name
	}
	return
}

/* 解密部分 */

type RawResponse struct {
	StartIndex int    `json:"startIndex"`
	EndIndex   int    `json:"endIndex"`
	S          string `json:"s"`
	D          string `json:"d"`
}

func DecodeResponse(body []byte) (data string, resp RawResponse) {
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Fatalln("响应解析成结构体失败：", err)
		return
	}
	iv := getMD5(resp.S + getRC4Enc())[resp.StartIndex:resp.EndIndex]
	data = DecryptAes128ECB(Base64Decode(resp.D), []byte(iv))
	return
}

func getMD5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func getRC4Enc() string {
	// ADW3123ASDQW22177EQWEQWEASD123ASWD14334Q123Q2
	c, err := rc4.NewCipher(key)
	if err != nil {
		log.Fatalln(err)
	}
	// 用于存放密文
	b := Base64Decode(secret)
	rc4Enc := make([]byte, len(b))
	c.XORKeyStream(rc4Enc, b)
	return string(rc4Enc)
}

func Base64Decode(s string) []byte {
	b, _ := base64.StdEncoding.DecodeString(s)
	return b
}

func DecryptAes128ECB(data, key []byte) string {
	var (
		block, _ = aes.NewCipher(key)
		dl       = len(data)
		result   = make([]byte, dl)
		bs       = block.BlockSize()
	)
	// ecb 需要分块解密
	for s, e := 0, bs; dl > s; s, e = s+bs, e+bs {
		block.Decrypt(result[s:e], data[s:e])
	}
	// 去除 padding
	for dl--; dl > -1; dl-- {
		if result[dl] > 31 {
			result = result[:dl+1]
			break
		}
	}
	return string(result)
}

type HttpBinIP struct {
	Origin string `json:"origin"`
}

func WhatsMyIP() {
	useProxy = true
	resp, err := utils.Get("https://httpbin.org/ip")
	useProxy = false
	if err != nil {
		log.Println("获取本机 ip 失败：", err)
		return
	}
	var ip HttpBinIP
	respBody, _ := io.ReadAll(resp.Body)
	if err = json.Unmarshal(respBody, &ip); err != nil {
		log.Println("获取本机ip解码返回结构出错：", err, string(respBody))
		return
	}
	println("本机公网 ip :", ip.Origin)
}

func Update(user ...string) (string,string) {
	account, password := "", "12345abc"
	switch len(user) {
	case 2:
		if user[1] != "" {
			password = user[1]
		}
		fallthrough
	case 1:
		account = user[0]
	}
	c := NewAStarVPNClient(account, password)
	if account == "" {
		c.Register()
		log.Println("注册完毕，正在获取代理列表")
	}
	pl := c.ProxyList()
	pls, _ := json.Marshal(pl)
	bnl := xray.BaseNodeConfigList{}
	_ = json.Unmarshal(pls, &bnl)
	ns := xray.GetSubscribeLinksByBaseNodeConfigList(bnl)
	return ns,string(pls)
}

func UpdateSubscribeFile(user ...string) {
  ns,_:= Update(user...)
	writeFile("app/astar_vpn/subscribe",ns)
}

func writeFile(filename, data string) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Println("写入文件错误：", filename, err)
		return
	}
	file.WriteString(data)
	defer file.Close()
}

func init() {
	utils.DefaultClient.Transport.(*http.Transport).Proxy = func(req *http.Request) (*url.URL, error) {
		if useProxy {
			// 使用终端代理
			return http.ProxyFromEnvironment(req)
		}
		return nil, nil
	}
}
