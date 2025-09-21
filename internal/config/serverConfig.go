package config

type ServerArg struct {
	Host     string
	Interval uint64
	File     string
	Restore  bool
	ConnStr  string
	Key      string
}
