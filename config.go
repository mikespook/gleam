package gleam

type Config struct {
	Brokers     []string
	Prefix      string
	ClientId    string
	StateUpdate int
	Tasks       map[string]byte
}
