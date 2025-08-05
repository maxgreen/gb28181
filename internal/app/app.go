package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/gowvp/gb28181/internal/conf"
	"github.com/ixugo/goddd/domain/version/versionapi"
	"github.com/ixugo/goddd/pkg/logger"
	"github.com/ixugo/goddd/pkg/server"
	"github.com/ixugo/goddd/pkg/system"
)

func Run(bc *conf.Bootstrap) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 以可执行文件所在目录为工作目录，防止以服务方式运行时，工作目录切换到其它位置
	bin, _ := os.Executable()
	if err := os.Chdir(filepath.Dir(bin)); err != nil {
		slog.Error("change work dir fail", "err", err)
	}

	log, clean := SetupLog(bc)
	defer clean()

	go setupZLM(ctx, bc.ConfigDir)

	// TODO: 异步发现 zlm 配置，有概率程序启动了，才找到 zlm 的秘钥，建议提前配置好秘钥
	go setupSecret(bc)
	// 如果需要执行表迁移，递增此版本号和表更新说明
	versionapi.DBVersion = "0.0.11"
	versionapi.DBRemark = "add stream proxy"

	handler, cleanUp, err := wireApp(bc, log)
	if err != nil {
		slog.Error("程序构建失败", "err", err)
		panic(err)
	}
	defer cleanUp()

	svc := server.New(handler,
		server.Port(strconv.Itoa(bc.Server.HTTP.Port)),
		server.ReadTimeout(bc.Server.HTTP.Timeout.Duration()),
		server.WriteTimeout(bc.Server.HTTP.Timeout.Duration()),
	)
	go svc.Start()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("服务启动成功 port:", bc.Server.HTTP.Port)

	select {
	case s := <-interrupt:
		slog.Info(`<-interrupt`, "signal", s.String())
	case err := <-svc.Notify():
		system.ErrPrintf("err: %s\n", err.Error())
		slog.Error(`<-server.Notify()`, "err", err)
	}
	cancel()
	if err := svc.Shutdown(); err != nil {
		slog.Error(`server.Shutdown()`, "err", err)
	}
}

// SetupLog 初始化日志
func SetupLog(bc *conf.Bootstrap) (*slog.Logger, func()) {
	logDir := filepath.Join(bc.ConfigDir, bc.Log.Dir)
	_ = os.MkdirAll(logDir, 0o755)
	return logger.SetupSlog(logger.Config{
		Dir:          logDir,                            // 日志地址
		Debug:        bc.Debug,                          // 服务级别Debug/Release
		MaxAge:       bc.Log.MaxAge.Duration(),          // 日志存储时间
		RotationTime: bc.Log.RotationTime.Duration(),    // 循环时间
		RotationSize: bc.Log.RotationSize * 1024 * 1024, // 循环大小
		Level:        bc.Log.Level,                      // 日志级别
	})
}

// 读取 config.ini 文件，通过正则表达式，获取 secret 的值
func getSecret(configDir string) (string, error) {
	for _, file := range []string{"zlm.ini", "config.ini"} {
		content, err := os.ReadFile(filepath.Join(configDir, file))
		if err != nil {
			continue
		}
		re := regexp.MustCompile(`secret=(\w+)`)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) < 2 {
			continue
		}
		return matches[1], nil
	}
	return "", fmt.Errorf("unknow")
}

func setupZLM(ctx context.Context, dir string) {
	// 检查是否在 Docker 环境中
	_, err := os.Stat("/.dockerenv")
	if !(err == nil || os.Getenv("NVR_STREAM") == "ZLM") {
		slog.Info("未在 Docker 环境中运行，跳过启动 zlm")
		return
	}

	// 检查 MediaServer 文件是否存在
	mediaServerPath := filepath.Join(system.Getwd(), "MediaServer")
	if _, err := os.Stat(mediaServerPath); os.IsNotExist(err) {
		slog.Info("MediaServer 文件不存在", "path", mediaServerPath)
		return
	}

	// 启动 MediaServer
	cmd := exec.CommandContext(ctx, "./MediaServer", "-s", "default.pem", "-c", filepath.Join(dir, "zlm.ini")) // nolint
	cmd.Dir = system.Getwd()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	for {
		slog.Info("MediaServer 启动中...")
		// 启动命令
		if err := cmd.Run(); err != nil {
			slog.Error("zlm 运行失败", "err", err)
			continue
		}
		time.Sleep(5 * time.Second)
	}
}

func setupSecret(bc *conf.Bootstrap) {
	for range 3 {
		secret, err := getSecret(bc.ConfigDir)
		if err == nil {
			slog.Info("发现 zlm 配置，已赋值，未回写配置文件", "secret", secret)
			bc.Media.Secret = secret
			return
		}
		time.Sleep(2 * time.Second)
		continue
	}
	slog.Warn("未发现 zlm 配置，请手动配置 zlm secret")
}
