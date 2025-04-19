package main

import (
	"expvar"
	"flag"
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

var (
	buildVersion = "0.0.1" // 构建版本号
	gitBranch    = "dev"   // git 分支
	gitHash      = "debug" // git 提交点哈希值
	release      string    // 发布模式 true/false
	buildTime    string    // 构建时间戳
)

// 自定义配置目录
var configDir = flag.String("conf", "./configs", "config directory, eg: -conf /configs/")

func getBuildRelease() bool {
	v, _ := strconv.ParseBool(release)
	return v
}

func main() {
	flag.Parse()
	// 以可执行文件所在目录为工作目录，防止以服务方式运行时，工作目录切换到其它位置
	bin, _ := os.Executable()
	if err := os.Chdir(filepath.Dir(bin)); err != nil {
		slog.Error("change dir error")
	}
	go setupZLM(*configDir)
	// 初始化配置
	var bc conf.Bootstrap
	// 获取配置目录绝对路径
	fileDir, _ := abs(*configDir)
	os.MkdirAll(fileDir, 0o755)
	filePath := filepath.Join(fileDir, "config.toml")
	configIsNotExistWrite(filePath)
	if err := conf.SetupConfig(&bc, filePath); err != nil {
		panic(err)
	}

	bc.Debug = !getBuildRelease()
	bc.BuildVersion = buildVersion
	bc.ConfigDir = fileDir
	bc.ConfigPath = filePath

	// 初始化日志
	logDir := filepath.Join(system.Getwd(), *configDir, bc.Log.Dir)
	if filepath.IsAbs(bc.Log.Dir) {
		logDir = bc.Log.Dir
	}
	log, clean := logger.SetupSlog(logger.Config{
		Dir:          logDir,                            // 日志地址
		Debug:        bc.Debug,                          // 服务级别Debug/Release
		MaxAge:       bc.Log.MaxAge.Duration(),          // 日志存储时间
		RotationTime: bc.Log.RotationTime.Duration(),    // 循环时间
		RotationSize: bc.Log.RotationSize * 1024 * 1024, // 循环大小
		Level:        bc.Log.Level,                      // 日志级别
	})
	{
		expvar.NewString("version").Set(buildVersion)
		expvar.NewString("git_branch").Set(gitBranch)
		expvar.NewString("git_hash").Set(gitHash)
		expvar.NewString("build_time").Set(buildTime)
		expvar.Publish("timestamp", expvar.Func(func() any {
			return time.Now().Format(time.DateTime)
		}))
	}

	go setupSecret(&bc)

	// 如果需要执行表迁移，递增此版本号和表更新说明
	versionapi.DBVersion = "0.0.10"
	versionapi.DBRemark = "add stream proxy"
	handler, cleanUp, err := wireApp(&bc, log)
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
	if err := svc.Shutdown(); err != nil {
		slog.Error(`server.Shutdown()`, "err", err)
	}

	defer clean()
}

func abs(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	bin, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(bin), path), nil
}

func configIsNotExistWrite(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := conf.WriteConfig(conf.DefaultConfig(), path); err != nil {
			system.ErrPrintf("WriteConfig", "err", err)
		}
	}
}

// 读取 config.ini 文件，通过正则表达式，获取 secret 的值
func getSecret(configDir string) (string, error) {
	for _, file := range []string{"zlm.ini", "config.ini"} {
		content, err := os.ReadFile(filepath.Join(system.Getwd(), configDir, file))
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

func setupZLM(dir string) {
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
	cmd := exec.Command("./MediaServer", "-s", "default.pem", "-c", filepath.Join(dir, "zlm.ini")) // nolint
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
		secret, err := getSecret(*configDir)
		if err == nil {
			slog.Info("发现 zlm 配置，已赋值，未回写配置文件", "secret", secret)
			bc.Media.Secret = secret
			return
		}
		time.Sleep(2 * time.Second)
		continue
	}
	slog.Info("未发现 zlm 配置，请检查 config.ini 文件")
}
