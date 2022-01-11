package vpn

type Connection struct {
	Name string

	LocalAddrs  []string            `vici:"local_addrs"`
	RemoteAddrs []string            `vici:"remote_addrs"`
	Local       *LocalAuth          `vici:"local"`
	Remote      *RemoteAuth         `vici:"remote"`
	Children    map[string]*ChildSA `vici:"children"`
	Version     int                 `vici:"version"`
	Proposals   []string            `vici:"proposals"`
	Keyingtries int                 `vici:"keyingtries"`
	//ReauthTime  int                 `vici:"reauth_time"`
}

type LocalAuth struct {
	Auth string `vici:"auth"`
	ID   string `vici:"id"`
}

type RemoteAuth struct {
	Auth string `vici:"auth"`
	ID   string `vici:"id"`
}

type ChildSA struct {
	LocalTS      []string `vici:"local_ts"`
	RemoteTS     []string `vici:"remote_ts"`
	ESPProposals []string `vici:"esp_proposals"`
	StartAction  string   `vici:"start_action"`
	//RekeyTime    int      `vici:"rekey_time"`
}

type Secret struct {
	ID     string   `vici:"id"`
	Type   string   `vici:"type"`
	Data   string   `vici:"data"`
	Owners []string `vici:"owners"`
}
