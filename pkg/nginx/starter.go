package nginx

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

func StartNginx(path string) {
	cmd := exec.Command("openresty", "-c", filepath.Join(path, "nginx.conf"))

	info, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("run nginx", "error", info)
		os.Exit(1)
	}
}

func StopNginx() {
	cmd := exec.Command("openresty", "-s", "stop")
	cmd.Run()
}
