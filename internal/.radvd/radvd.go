package radvd

type Interface struct {
	Instance             uint32    `json:"instance"`
	Name                 string    `json:"name"`
	AdvSendAdvert        bool      `json:"adv_send_advert"`
	MinRtrAdvInterval    uint32    `json:"min_rtr_adv_interval"`
	MaxRtrAdvInterval    uint32    `json:"max_rtr_adv_interval"`
	AdvManagedFlag       bool      `json:"adv_managed_flag"`
	AdvOtherConfigFlag   bool      `json:"adv_other_config_flag"`
	AdvDefaultLifetime   uint32    `json:"adv_default_lifetime"`
	AdvDefaultPreference string    `json:"adv_default_preference"`
	Prefixes             []*Prefix `json:"prefixes"`
	Rdnss                []*RDNSS  `json:"rdnss"`
	Routes               []*Route  `json:"routes"`
	Clients              []string  `json:"clients"`
}

type Prefix struct {
	Prefix           string `json:"prefix"`
	AdvOnLink        bool   `json:"adv_on_link"`
	AdvAutonomous    bool   `json:"adv_autonomous"`
	AdvRouterAddr    bool   `json:"adv_router_addr"`
	AdvValidLifetime uint32 `json:"adv_valid_lifetime"`
}

type RDNSS struct {
	Address          string `json:"address"`
	AdvRdnssLifetime uint32 `json:"adv_rdnss_lifetime"`
}

type Route struct {
	Route              string `json:"route"`
	AdvRouteLifetime   uint32 `json:"adv_route_lifetime"`
	AdvRoutePreference string `json:"adv_route_preference"`
}
