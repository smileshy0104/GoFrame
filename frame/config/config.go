package config

import (
	"flag"
	newlogger "frame/log"
	"github.com/BurntSushi/toml"
	"os"
)

// Conf 是一个全局变量，指向 FrameConfig 结构体实例，用于存储项目的配置信息。
// 它在程序初始化时通过 loadToml 函数加载配置文件内容。
var Conf = &FrameConfig{
	logger: newlogger.Default(),
}

// FrameConfig 定义了项目的配置结构，包含以下字段：
// - logger：日志记录器实例，用于记录日志信息。
// - Log：日志相关的配置项，存储为键值对形式。
// - Pool：连接池相关的配置项，存储为键值对形式。
// - Template：模板相关的配置项，存储为键值对形式。
type FrameConfig struct {
	logger   *newlogger.Logger
	Log      map[string]any
	Pool     map[string]any
	Template map[string]any
}

// init 函数在程序启动时自动调用，用于初始化全局配置 Conf。
// 它会调用 loadToml 函数加载配置文件。
func init() {
	loadToml()
}

// loadToml 用于加载 TOML 格式的配置文件，并将其解析到全局配置 Conf 中。
// 参数说明：
// - 配置文件路径通过命令行参数 "-conf" 指定，默认值为 "conf/app.toml"。
// 返回值：
// - 无返回值，但会在加载失败时记录日志并终止加载流程。
func loadToml() {
	// 定义命令行参数 "-conf"，用于指定配置文件路径，默认值为 "conf/app.toml"。
	configFile := flag.String("conf", "conf/app.toml", "app config file")
	flag.Parse()

	// 检查配置文件是否存在，如果不存在则记录日志并退出函数。
	if _, err := os.Stat(*configFile); err != nil {
		Conf.logger.Info("conf/app.toml file not load，because not exist")
		return
	}

	// 解析配置文件内容到 Conf 全局变量中，如果解析失败则记录日志并退出函数。
	_, err := toml.DecodeFile(*configFile, Conf)
	if err != nil {
		Conf.logger.Info("conf/app.toml decode fail check format")
		return
	}
}
