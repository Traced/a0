package xray

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Traced/a0/utils"
)

type (
	BaseNodeConfigList []BaseNodeConfig
	BaseNodeConfig     struct {
		Name       string `json:"name"`
		Password   string `json:"sd"`
		Pr         uint8  `json:"pr"`
		Address    string `json:"d"`
		Type       string `json:"t"`
		TLS        string `json:"tls"`
		UUID       string `json:"uuid"`
		Port       string `json:"o"`
		Flow       string `json:"lk"`
		Level      int    `json:"l"`
		Network    string `json:"csxy"`
		Encryption string
	}
	NodeUser struct {
		ID         string `json:"id"`
		AlterID    int    `json:"alterId"`
		Security   string `json:"security"`
		Encryption string `json:"encryption"`
		Flow       string `json:"flow"`
	}
	Node struct {
		Address string     `json:"address"`
		Port    int        `json:"port"`
		Users   []NodeUser `json:"users"`
	}
	Outbounds       []Outbound
	OutboundSetting struct {
		Vnext   []Node `json:"vnext,omitempty"`
		Servers []struct {
			Address  string `json:"address"`
			Password string `json:"password"`
			Port     int    `json:"port"`
		} `json:"servers,omitempty"`
		Response struct {
			Type string `json:"type"`
		}
	}
	Outbound struct {
		Tag            string          `json:"tag"`
		Protocol       string          `json:"protocol"`
		Settings       OutboundSetting `json:"settings"`
		StreamSettings *struct {
			Network      string `json:"network"`
			Security     string `json:"security"`
			XTlsSettings struct {
				AllowInsecure bool `json:"allowInsecure"`
			} `json:"xtlsSettings"`
		} `json:"streamSettings,omitempty"`
		Mux *struct {
			Enabled     bool `json:"enabled"`
			Concurrency int  `json:"concurrency"`
		} `json:"mux,omitempty"`
	}
	Config struct {
		Log struct {
			Loglevel string `json:"loglevel"`
		} `json:"log"`
		Inbounds []struct {
			Tag      string `json:"tag"`
			Port     int    `json:"port"`
			Listen   string `json:"listen"`
			Protocol string `json:"protocol"`
			Sniffing struct {
				Enabled      bool     `json:"enabled"`
				DestOverride []string `json:"destOverride"`
			} `json:"sniffing"`
			Settings struct {
				Udp              bool `json:"udp"`
				AllowTransparent bool `json:"allowTransparent"`
			} `json:"settings"`
		} `json:"inbounds"`
		Outbounds []Outbound `json:"outbounds"`
		Routing   struct {
			DomainStrategy string `json:"domainStrategy"`
			Rules          []struct {
				Type        string   `json:"type"`
				InboundTag  []string `json:"inboundTag,omitempty"`
				OutboundTag string   `json:"outboundTag"`
				Domain      []string `json:"domain,omitempty"`
				Ip          []string `json:"ip,omitempty"`
			} `json:"rules"`
		} `json:"routing"`
	}
)

var (
	template       = `{"log":{"loglevel":"warning"},"inbounds":[{"tag":"http","port":2021,"listen":"127.0.0.1","protocol":"http","sniffing":{"enabled":true,"destOverride":["http","tls"]},"settings":{"udp":false,"allowTransparent":false}}],"outbounds":[{"tag":"proxy","protocol":"","settings":{"servers":[{"address":"","password":"","port":0}],"vnext":[{"address":"","port":8001,"users":[{"id":"","alterId":0,"security":"auto","encryption":"none","flow":""}]}]},"streamSettings":{"network":"tcp","security":"tls","xtlsSettings":{"allowInsecure":false}},"mux":{"enabled":false,"concurrency":-1}},{"tag":"direct","protocol":"freedom","settings":{}},{"tag":"block","protocol":"blackhole","settings":{"response":{"type":"http"}}}],"routing":{"domainStrategy":"IPIfNonMatch","rules":[{"type":"field","inboundTag":["api"],"outboundTag":"api"},{"type":"field","outboundTag":"direct","domain":["geosite:cn"]},{"type":"field","outboundTag":"direct","ip":["geoip:cn"]}]}}`
	configTemplate = new(Config)
	_              = json.Unmarshal([]byte(template), configTemplate)
)

func GetConfigTemplate() *Config {
	return configTemplate
}

func GetSubscribeLinksByBaseNodeConfigList(bnl BaseNodeConfigList) string {
	var buf strings.Builder
	for _, bn := range bnl {
		bn.Password = strings.ReplaceAll(bn.Password, "\t", "\\t")
		nl := GetSubscribeLinkByBaseNodeConfig(bn)
		buf.WriteString(nl)
		buf.WriteByte('\n')
	}
	return utils.Base64EncodeString(buf.String())
}

func GetSubscribeLinkByBaseNodeConfig(bn BaseNodeConfig) string {
	link := ""
	remark := url.QueryEscape(bn.Name)
	switch bn.Type {
	case "vless":
		link = fmt.Sprintf("vless://%s@%s:%s?security=%s&type=%s&flow=%s#%s",
			bn.UUID,
			bn.Address,
			bn.Port,
			strings.ToLower(bn.TLS),
			bn.Network,
			strings.ToLower(bn.Flow),
			remark,
		)
	case "trojan":
		link = fmt.Sprintf("trojan://%s@%s:%s#%s", bn.Password, bn.Address, bn.Port, remark)
	}
	return link
}

func GetSubscribeLinkStringByOutbound(o Outbound, remark string) string {
	link := ""
	switch o.Protocol {
	case "vless":
		link = fmt.Sprintf("vless://%s@%s:%d?encryption=%s&security=%s&flow=%s&aid=%d&ps=%s",
			o.Settings.Vnext[0].Users[0].ID,
			o.Settings.Vnext[0].Address,
			o.Settings.Vnext[0].Port,
			o.Settings.Vnext[0].Users[0].Encryption,
			o.StreamSettings.Security,
			o.Settings.Vnext[0].Users[0].Flow,
			o.Settings.Vnext[0].Users[0].AlterID,
			remark,
		)
	case "trojan":
		link = fmt.Sprintf("trojan://%s@%s:%d",
			o.Settings.Servers[0].Password,
			o.Settings.Servers[0].Address,
			o.Settings.Servers[0].Port,
		)
	}
	return link
}

func GetConfigByBaseNodeConfig(bn BaseNodeConfig) Config {
	var (
		ct      = *configTemplate
		port, _ = strconv.Atoi(bn.Port)
	)
	ct.Outbounds[0].Protocol = bn.Type
	if bn.Network == "" {
		bn.Network = "tcp"
	}
	if bn.TLS == "" {
		bn.TLS = "tls"
	}
	ct.Outbounds[0].StreamSettings.Network = bn.Network
	ct.Outbounds[0].StreamSettings.Security = bn.TLS
	// 不同协议
	switch bn.Type {
	case "vless":
		ct.Outbounds[0].Settings.Vnext[0].Address = bn.Address
		ct.Outbounds[0].Settings.Vnext[0].Port = port
		ct.Outbounds[0].Settings.Vnext[0].Users[0].AlterID = bn.Level
		ct.Outbounds[0].Settings.Vnext[0].Users[0].Flow = bn.Flow
		ct.Outbounds[0].Settings.Vnext[0].Users[0].ID = bn.UUID
	case "trojan":
		ct.Outbounds[0].Settings.Servers[0].Address = bn.Address
		ct.Outbounds[0].Settings.Servers[0].Password = bn.Password
		ct.Outbounds[0].Settings.Servers[0].Port = port
	}
	return ct
}
