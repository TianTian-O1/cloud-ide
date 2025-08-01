package conf

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"strconv"

	"github.com/mangohow/cloud-ide/pkg/conf"
	"github.com/spf13/viper"
)

var (
	ServerConfig conf.ServerConf
	MysqlConfig  conf.MysqlConf
	RedisConfig  conf.RedisConf
	LoggerConfig conf.LoggerConf
	GrpcConfig   conf.GrpcConf
	EmailConfig  conf.EmailConf
	OAuthConfig  conf.OAuthConf
)

func LoadConf() error {
	viper.SetConfigName("webserver")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/app/config")  // 添加Kubernetes挂载路径
	
	// 读取配置文件，如果不存在则忽略错误
	err := viper.ReadInConfig()
	if err != nil {
		// 配置文件不存在时，使用默认配置
		viper.Set("server.host", "0.0.0.0")
		viper.Set("server.port", 8088)
		viper.Set("server.name", "cloud-ide-webserver")
		viper.Set("server.mode", "release")
		
		viper.Set("logger.level", "info")
		viper.Set("logger.toFile", false)
		
		viper.Set("email.enabled", false)
		
		viper.Set("redis.addr", "redis-svc.cloud-ide.svc.cluster.local:6379")  // 设置默认Redis地址
		viper.Set("grpc.addr", "cloud-ide-control-plane-svc:6387")
	}

	initServerConf()
	initMysqlConf()
	initRedisConf()
	initLogConf()
	initGrpcConf()
	initEmailConf()
	initOAuthConf()

	parseFlags()

	return nil
}

func initServerConf() {
	ServerConfig = conf.ServerConf{
		Host: viper.GetString("server.host"),
		Port: viper.GetInt("server.port"),
		Name: viper.GetString("server.name"),
		Mode: viper.GetString("server.mode"),
	}
}

func initMysqlConf() {
	// 优先从环境变量读取数据库配置
	dbHost := viper.GetString("mysql.host")
	dbPort := viper.GetString("mysql.port")
	dbUser := viper.GetString("mysql.user")
	dbPassword := viper.GetString("mysql.password")
	dbName := viper.GetString("mysql.database")
	
	// 如果配置文件中没有MySQL配置，从环境变量构建DSN
	dataSourceName := viper.GetString("mysql.dataSourceName")
	if dataSourceName == "" && dbHost == "" {
		// 从deployment环境变量中读取
		dbHost = os.Getenv("DB_HOST")
		dbPort = os.Getenv("DB_PORT")
		dbUser = os.Getenv("DB_USER")
		dbPassword = os.Getenv("DB_PASSWORD")
		dbName = os.Getenv("DB_NAME")
		
		if dbHost != "" {
			if dbPort == "" {
				dbPort = "3306"
			}
			dataSourceName = dbUser + ":" + dbPassword + "@(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8mb4&parseTime=true&loc=Local"
		}
	}
	
	MysqlConfig = conf.MysqlConf{
		DataSourceName: dataSourceName,
		MaxOpenConns:   viper.GetInt("mysql.maxOpenConns"),
		MaxIdleConns:   viper.GetInt("mysql.maxIdleConns"),
	}
	
	// 设置默认值
	if MysqlConfig.MaxOpenConns == 0 {
		MysqlConfig.MaxOpenConns = 10
	}
	if MysqlConfig.MaxIdleConns == 0 {
		MysqlConfig.MaxIdleConns = 5
	}
}

func initLogConf() {
	LoggerConfig = conf.LoggerConf{
		Level:       viper.GetString("logger.level"),
		FilePath:    viper.GetString("logger.filePath"),
		FileName:    viper.GetString("logger.fileName"),
		MaxFileSize: viper.GetUint64("logger.maxFileSize"),
		ToFile:      viper.GetBool("logger.toFile"),
	}
}

func initRedisConf() {
	RedisConfig = conf.RedisConf{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetUint32("redis.db"),
		PoolSize:     viper.GetUint32("redis.poolSize"),
		MinIdleConns: viper.GetUint32("redis.minIdleConns"),
	}
	
	// 从环境变量覆盖Redis配置
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		RedisConfig.Addr = addr
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		RedisConfig.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.ParseUint(db, 10, 32); err == nil {
			RedisConfig.DB = uint32(d)
		}
	}
	if poolSize := os.Getenv("REDIS_POOL_SIZE"); poolSize != "" {
		if p, err := strconv.ParseUint(poolSize, 10, 32); err == nil {
			RedisConfig.PoolSize = uint32(p)
		}
	}
	if minIdleConns := os.Getenv("REDIS_MIN_IDLE_CONNS"); minIdleConns != "" {
		if m, err := strconv.ParseUint(minIdleConns, 10, 32); err == nil {
			RedisConfig.MinIdleConns = uint32(m)
		}
	}
	
	// 设置默认值
	if RedisConfig.PoolSize == 0 {
		RedisConfig.PoolSize = 10
	}
	if RedisConfig.MinIdleConns == 0 {
		RedisConfig.MinIdleConns = 5
	}
}

func initGrpcConf() {
	// 优先从环境变量读取GRPC地址
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = viper.GetString("grpc.addr")
	}
	
	GrpcConfig = conf.GrpcConf{Addr: grpcAddr}
	
	// 添加调试日志
	fmt.Printf("[DEBUG] GRPC Config: %s\n", grpcAddr)
}

func initEmailConf() {
	// 优先从环境变量读取邮件配置
	emailEnabled := os.Getenv("EMAIL_ENABLED")
	
	EmailConfig = conf.EmailConf{
		Enabled:     viper.GetBool("email.enabled"),
		Host:        viper.GetString("email.host"),
		Port:        viper.GetUint32("email.port"),
		SenderEmail: viper.GetString("email.senderEmail"),
		AuthCode:    viper.GetString("email.authCode"),
	}
	
	// 如果环境变量设置了EMAIL_ENABLED，使用环境变量的值
	if emailEnabled != "" {
		EmailConfig.Enabled = strings.ToLower(emailEnabled) == "true"
	}
	
	// 从环境变量覆盖其他邮件配置
	if host := os.Getenv("EMAIL_HOST"); host != "" {
		EmailConfig.Host = host
	}
	if port := os.Getenv("EMAIL_PORT"); port != "" {
		if p, err := strconv.ParseUint(port, 10, 32); err == nil {
			EmailConfig.Port = uint32(p)
		}
	}
	if sender := os.Getenv("EMAIL_SENDER"); sender != "" {
		EmailConfig.SenderEmail = sender
	}
	if authCode := os.Getenv("EMAIL_AUTH_CODE"); authCode != "" {
		EmailConfig.AuthCode = authCode
	}
}

func initOAuthConf() {
	OAuthConfig = conf.OAuthConf{
		LinuxDoClientID:     viper.GetString("oauth.linuxdo.client_id"),
		LinuxDoClientSecret: viper.GetString("oauth.linuxdo.client_secret"),
		LinuxDoRedirectURL:  viper.GetString("oauth.linuxdo.redirect_url"),
		LinuxDoBaseURL:      viper.GetString("oauth.linuxdo.base_url"),
	}
	
	// 从环境变量覆盖OAuth配置
	if clientID := os.Getenv("LINUXDO_CLIENT_ID"); clientID != "" {
		OAuthConfig.LinuxDoClientID = clientID
	}
	if clientSecret := os.Getenv("LINUXDO_CLIENT_SECRET"); clientSecret != "" {
		OAuthConfig.LinuxDoClientSecret = clientSecret
	}
	if redirectURL := os.Getenv("LINUXDO_REDIRECT_URL"); redirectURL != "" {
		OAuthConfig.LinuxDoRedirectURL = redirectURL
	}
	if baseURL := os.Getenv("LINUXDO_BASE_URL"); baseURL != "" {
		OAuthConfig.LinuxDoBaseURL = baseURL
	}
	
	// 设置默认值
	if OAuthConfig.LinuxDoBaseURL == "" {
		OAuthConfig.LinuxDoBaseURL = "https://connect.linux.do"
	}
	if OAuthConfig.LinuxDoRedirectURL == "" {
		OAuthConfig.LinuxDoRedirectURL = "https://tiantianai.co/auth/oauth/linuxdo/callback"
	}
}

// 解析命令行参数
func parseFlags() {
	var (
		mode           string
		port           int
		dataSourceName string
		logLevel       string
		email          string
		emailHost      string
		emailPort      int
		senderEmail    string
		authCode       string
		grpcAddr       string
	)

	flag.StringVar(&mode, "mode", "", "specify server running mode [dev, release]")
	flag.IntVar(&port, "port", -1, "specify server listen port")
	flag.StringVar(&dataSourceName, "mysql-datasource", "", "specify mysql datasource eg. user:password@(svc:port)/database?charset=utf8mb4&parseTime=true&loc=Local")
	flag.StringVar(&logLevel, "log-level", "", "specify log level [debug, info, warn, error]")
	flag.StringVar(&email, "email-enabled", "", "enable email register [enabled, disabled]")
	flag.StringVar(&emailHost, "email-host", "", "specify email host if email is enabled")
	flag.IntVar(&emailPort, "email-port", -1, "specify email port if email is enabled")
	flag.StringVar(&senderEmail, "email-sender", "", "specify sender email if email is enabled")
	flag.StringVar(&authCode, "email-authcode", "", "specify email auth code if email is enabled")
	flag.StringVar(&grpcAddr, "grpc-addr", "", "specify control plane grpc addr eg:cloud-ide-control-plane-svc:6387")
	flag.Parse()

	setString(&ServerConfig.Mode, &mode)
	setString(&EmailConfig.SenderEmail, &senderEmail)
	setString(&EmailConfig.AuthCode, &authCode)
	// 添加调试日志
	fmt.Printf("[DEBUG] Command line grpcAddr: '%s'\n", grpcAddr)
	fmt.Printf("[DEBUG] Before parseFlags GrpcConfig.Addr: '%s'\n", GrpcConfig.Addr)
	
	setString(&GrpcConfig.Addr, &grpcAddr)
	
	fmt.Printf("[DEBUG] After parseFlags GrpcConfig.Addr: '%s'\n", GrpcConfig.Addr)
	
	setString(&MysqlConfig.DataSourceName, &dataSourceName)
	setString(&LoggerConfig.Level, &logLevel)
	setString(&EmailConfig.Host, &emailHost)
	if port != -1 {
		ServerConfig.Port = port
	}
	if emailPort != -1 {
		EmailConfig.Port = uint32(emailPort)
	}
	switch strings.ToLower(email) {
	case "enabled":
		EmailConfig.Enabled = true
	case "disabled":
		EmailConfig.Enabled = false
	}
}

func setString(dst *string, src *string) {
	if *src != "" {
		*dst = *src
	}
}
