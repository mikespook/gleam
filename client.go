package gleam

import "github.com/coreos/go-etcd/etcd"

type Client struct {
	*etcd.Client
}

func NewClient(machines []string, cert, key, ca string) (client *Client, err error) {
	client = &Client{}
	if cert != "" && key != "" && ca != "" {
		if client.Client, err = etcd.NewTLSClient(machines, cert, key, ca); err != nil {
			return
		}
	} else {
		client.Client = etcd.NewClient(machines)
	}
	return
}

func (client *Client) List(dir string) (map[string]string, error) {
	r, err := client.Get(dir, true, false)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, r.Node.Nodes.Len())
	for _, n := range r.Node.Nodes {
		m[n.Key] = n.Value
	}
	return m, nil
}

func (client *Client) Call(key, funcname string, data interface{}) error {
	strdata, err := unmarshal(&Func{funcname, data})
	if err != nil {
		return err
	}
	_, err = client.Set(key, strdata, 0)
	return err
}
