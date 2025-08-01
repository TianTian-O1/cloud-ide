package tmpl

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type Config struct {
	WorkerProcess     int
	WorkerConnections int
	SharedDictSize    string
	NginxLuaPath      string
	Debug             bool
	Token             string
	ServerCrt         string
	ServerKey         string
	WebServiceName    string
	WebPort           int
}

func ApplyNginxConf(cfg *Config, ngxPath string) error {
	err := applyTmpl(cfg, ngxPath)
	if err != nil {
		return err
	}

	return moveLuaFiles(ngxPath)
}

func applyTmpl(cfg *Config, ngxPath string) error {
	tmpl, err := template.ParseFiles("nginx/nginx.tmpl")
	if err != nil {
		slog.Error("parse nginx file", "error", err)
		return err
	}

	filename := filepath.Join(ngxPath, "nginx.conf")
	// 先删除
	os.Remove(filename)

	file, err := os.Create(filename)
	if err != nil && err != os.ErrExist {
		slog.Error("create nginx.conf", "error", err)
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, cfg)
	if err != nil {
		slog.Error("execute tmpl", "error", err)
	}

	return err
}

func moveLuaFiles(ngxPath string) error {
	cmd := exec.Command("cp", "-r", "nginx/lua", filepath.Join(ngxPath, "lua"))
	res, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("move lua files", "error", res)
		return err
	}

	return nil
}
