package yun_ti

type (
	baseResponse struct {
		BackendUrl  string `json:"backend_url"`
		Code        int    `json:"code"`
		FrontendUrl string `json:"frontend_url"`
		InviteCode  string `json:"invite_code"`
		InviteUrl   string `json:"invite_url"`
		Message     string `json:"message"`
		Now         string `json:"now"`
		PayUrl      string `json:"pay_url"`
		SupportUrl  string `json:"support_url"`
		TerminalId  string `json:"terminal_id"`
		Token       string `json:"token"`
		TransferUrl string `json:"transfer_url"`
		UserId      int    `json:"user_id"`
		UserInfo    *struct {
			DateAdded    string `json:"date_added"`
			Deadline     string `json:"deadline"`
			DeadlineDate string `json:"deadline_date"`
			Email        string `json:"email"`
			Enable       int    `json:"enable"`
			Passwd       string `json:"passwd"`
			Prefix       string `json:"prefix"`
			Role         string `json:"role"`
			TerminalIds  string `json:"terminal_ids"`
			UserId       int    `json:"user_id"`
		} `json:"user_info"`
	}
	nodeResponse struct {
		baseResponse
		KcpInfo *struct {
			Crypt       string `json:"crypt"`
			Datashard   int    `json:"datashard"`
			Dscp        int    `json:"dscp"`
			KcpPort     int    `json:"kcp_port"`
			Key         string `json:"key"`
			Mode        string `json:"mode"`
			Mtu         int    `json:"mtu"`
			Parityshard int    `json:"parityshard"`
			Quiet       bool   `json:"quiet"`
			Rcvwnd      int    `json:"rcvwnd"`
			Sndwnd      int    `json:"sndwnd"`
		} `json:"kcp_info"`
		Nodes []*struct {
			Area           string `json:"area"`
			Domain         string `json:"domain"`
			Http2QrcodeUrl string `json:"http2_qrcode_url"`
			Ip             string `json:"ip"`
			Method         string `json:"method"`
			NaiveUrl       string `json:"naive_url"`
			Name           string `json:"name"`
			NameEn         string `json:"name_en"`
			NodeId         int    `json:"node_id"`
			NodeName       string `json:"node_name"`
			Obfs           string `json:"obfs"`
			ObfsParam      string `json:"obfs_param"`
			Passwd         string `json:"passwd"`
			Port           string `json:"port"`
			Protocol       string `json:"protocol"`
			ProtocolParam  string `json:"protocol_param"`
			QrCodeUrl      string `json:"qr_code_url"`
			SsrJson        struct {
				LocalAddress  string `json:"local_address"`
				LocalPort     int    `json:"local_port"`
				Method        string `json:"method"`
				Obfs          string `json:"obfs"`
				ObfsParam     string `json:"obfs_param"`
				Password      string `json:"password"`
				Protocol      string `json:"protocol"`
				ProtocolParam string `json:"protocol_param"`
				Server        string `json:"server"`
				ServerPort    string `json:"server_port"`
				Timeout       int    `json:"timeout"`
			} `json:"ssr_json"`
			SupportKcp   bool `json:"support_kcp"`
			SupportNaive bool `json:"support_naive"`
		} `json:"nodes"`
	}
)
