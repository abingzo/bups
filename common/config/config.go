package config

import (
	"github.com/BurntSushi/toml"
	"io"
)

type AutoGenerated struct {
	Main struct {
		Install  []string `toml:"install"`
		LoppTime int      `toml:"lopp_time"`
	} `toml:"main"`
	Plugin map[string]map[string]map[string]interface{} `toml:"plugin"`
	// 插件获取配置相关
	pluginName string
	scope string
}

func (a *AutoGenerated) SetPluginName(name string) {
	a.pluginName = name
}

func (a *AutoGenerated) SetPluginScope(scope string) {
	a.scope = scope
}

func (a *AutoGenerated) PluginGetData(key string) interface{} {
	return a.Plugin[a.pluginName][a.scope][key]
}

// 插件获取自身的对应的配置

func Read(reader io.Reader) *AutoGenerated {
	ag := &AutoGenerated{}
	_, err := toml.DecodeReader(reader,ag)
	if err != nil {
		panic(err)
	}
	return ag
}

func Write(writer io.Writer,cfg *AutoGenerated) error {
	return toml.NewEncoder(writer).Encode(cfg)
}