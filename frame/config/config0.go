package config

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/spf13/viper"
)

var AppConf *viper.Viper

// 初始化配置文件
func InitConfig(f embed.FS) {
	AppConf = viper.New()
	//添加根路径
	dir := "config"
	// 获取目录的文件名
	dirEntries, err := f.ReadDir(dir)
	if err != nil {
		fmt.Println("错误", err.Error())
	}
	for _, de := range dirEntries {
		if !de.IsDir() {
			file, _ := f.ReadFile(dir + "/" + de.Name())
			// 如果你的配置文件没有写扩展名，那么这里需要声明你的配置文件属于什么格式
			AppConf.SetConfigType("yaml")
			_ = AppConf.MergeConfig(bytes.NewBuffer(file))
		}
	}

}
