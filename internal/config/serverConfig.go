package config

type ServerArg struct {
	Host     string `arg:"-a,env:ADDRESS" help:"адрес сервера" default:"localhost:8080"`
	Interval uint64 `arg:"-i,env:STORE_INTERVAL" help:"store interval" default:"300"`
	File     string `arg:"-f,env:FILE_STORAGE_PATH" help:"file name" default:"metrics.json"`
	Restore  bool   `arg:"-r,env:RESTORE" help:"load history" default:"true"`
	ConnStr  string `arg:"-d,env:DATABASE_DSN" help:"db address"`
	Key      string `arg:"-k,env:KEY" help:"key"`
}
