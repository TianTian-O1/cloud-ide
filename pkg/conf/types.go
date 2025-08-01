package conf

type ServerConf struct {
	Host string
	Port int
	Name string
	Mode string
}

type MysqlConf struct {
	DataSourceName string
	MaxOpenConns   int
	MaxIdleConns   int
}

type RedisConf struct {
	Addr         string
	PoolSize     uint32
	MinIdleConns uint32
	Password     string
	DB           uint32
}

type LoggerConf struct {
	Level       string
	FilePath    string
	FileName    string
	MaxFileSize uint64
	ToFile      bool
}

type EmailConf struct {
	Enabled     bool
	Host        string
	Port        uint32
	SenderEmail string
	AuthCode    string
}

type GrpcConf struct {
	Addr string
}

type OAuthConf struct {
	LinuxDoClientID     string
	LinuxDoClientSecret string
	LinuxDoRedirectURL  string
	LinuxDoBaseURL      string
}
