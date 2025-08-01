package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/mangohow/cloud-ide/pkg/nginx"
	"github.com/mangohow/cloud-ide/pkg/tmpl"
	_ "go.uber.org/automaxprocs"
)

var (
	workerProcess     string
	workerConnections int
	sharedDictSize    string
	nginxConfPath     string
	debug             string
	token             string
	serverCrt         string
	serverKey         string
	webSvcName        string
	webPort           int
)

func main() {
	gomaxprocs := runtime.GOMAXPROCS(0)
	runtime.GOMAXPROCS(1)

	// 解析flags
	cfg, err := parseFlags(gomaxprocs)
	if err != nil {
		os.Exit(1)
	}

	// 生成nginx配置文件
	err = tmpl.ApplyNginxConf(cfg, nginxConfPath)
	if err != nil {
		os.Exit(1)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	go signalHandler(cancelFunc)

	go func() {
		<-ctx.Done()
		nginx.StopNginx()
	}()

	// 启动nginx
	nginx.StartNginx(nginxConfPath)
}

func parseFlags(gomaxprocs int) (*tmpl.Config, error) {
	flag.StringVar(&workerProcess, "workers", "1", "specify nginx worker_processes")
	flag.IntVar(&workerConnections, "conns-per-worker", 1024, "specify nginx worker_connections")
	flag.StringVar(&sharedDictSize, "shared-dict-size", "16m", "specify nginx shared dict size")
	flag.StringVar(&nginxConfPath, "nginx-conf-path", "/usr/local/openresty/nginx/conf", "specify nginx shared dict size")
	flag.StringVar(&debug, "debug", "disabled", "specify debug mode")
	flag.StringVar(&token, "endpoint-token", "", "specify endpoint token")
	flag.StringVar(&serverCrt, "server-crt", "", "specify ssl certificate")
	flag.StringVar(&serverKey, "server-key", "", "specify ssl certificate key")
	flag.StringVar(&webSvcName, "web-service-name", "cloud-ide-web-svc.cloud-ide.svc.cluster.local", "specify the service of the web to reverse proxy, fully qualified domain names must be written")
	flag.IntVar(&webPort, "web-port", 8088, "specify the port of the web to reverse proxy")
	flag.Parse()

	cfg := &tmpl.Config{}

	// TLS证书解析验证
	if _, err := tls.LoadX509KeyPair(serverCrt, serverKey); err != nil {
		slog.Error("ssl config", "error", err)
		return nil, err
	}
	cfg.ServerCrt = serverCrt
	cfg.ServerKey = serverKey

	if workerProcess == "auto" {
		cfg.WorkerProcess = gomaxprocs
	} else {
		n, err := strconv.Atoi(workerProcess)
		if err != nil {
			slog.Error("set workers", "error", "must be a number")
			return nil, err
		}
		cfg.WorkerProcess = n
	}

	if workerConnections < 1024 {
		workerConnections = 1024
	}
	cfg.WorkerConnections = workerConnections

	// 共享内存 单位为m,k 必须大于12k
	c := sharedDictSize[len(sharedDictSize)-1]
	if c != 'm' && c != 'k' {
		slog.Error("set shared dict size", "error", "unit must be 'm' or 'k'")
		return nil, errors.New("unit invalid")
	}
	sz, err := strconv.Atoi(sharedDictSize[:len(sharedDictSize)-1])
	if err != nil {
		slog.Error("set shared dict size", "error", "shared dict size invalid")
		return nil, errors.New("unit invalid")
	}

	if c == 'k' && sz < 12 {
		sharedDictSize = "12k"
	}
	cfg.SharedDictSize = sharedDictSize

	cfg.NginxLuaPath = filepath.Join(nginxConfPath, "lua")

	if debug == "enabled" {
		cfg.Debug = true
	}

	if token == "" {
		slog.Error("must specify endpoint token")
		return nil, errors.New("must specify endpoint token")
	}
	cfg.Token = token

	cfg.WebServiceName = webSvcName
	cfg.WebPort = webPort

	return cfg, nil
}

func signalHandler(exit func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
	exit()
}
