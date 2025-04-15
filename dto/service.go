package dto

type NewServiceReq struct {
	Name          string   `json:"name"`
	TCP           bool     `json:"tcp"`
	UDP           bool     `json:"udp"`
	HTTP          bool     `json:"http"`
	PortBlackList []string `json:"port_black_list"`
	PortWitheList []string `json:"port_white_list"`
	Options       []string `json:"options"`
}

type NewServiceRes struct {
	Id   string `json:"id"`
	HTTP string `json:"http_port"`
	TCP  string `json:"tcp_port"`
	UDP  string `json:"udp_port"`
}
