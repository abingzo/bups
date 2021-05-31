package conf

import (
	"github.com/BurntSushi/toml"
	this "github.com/mengzushan/bups/common/error"
	"github.com/mengzushan/bups/common/logger"
	"github.com/mengzushan/bups/utils"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type ConfNum int

const (
	ConfModelDev ConfNum = 1
	ConfModelPro ConfNum = 2
	ConfModelPre ConfNum = 3
)

var ConfModel ConfNum

// 测试使用的环境变量
var TestOnFilePath string

type AutoGenerated struct {
	CloudAPI     string `toml:"cloud_api"`
	SaveName     string `toml:"save_name"`
	SaveTime     int    `toml:"save_time"`
	SaveTimePass int    `toml:"save_time_pass"`
	Bucket       struct {
		BucketURL string `toml:"bucket_url"`
		Secretid  string `toml:"secretid"`
		Secretkey string `toml:"secretkey"`
		Token     string `toml:"token"`
	} `toml:"bucket"`
	Database struct {
		Ipaddr     string `toml:"ipaddr"`
		Port       string `toml:"port"`
		UserName   string `toml:"user_name"`
		UserPasswd string `toml:"user_passwd"`
		DbName     string `toml:"db_name"`
		DbName2    string `toml:"db_name2"`
	} `toml:"database"`
	Local struct {
		Web    string `toml:"web"`
		Static string `toml:"static"`
		Log    string `toml:"log"`
	} `toml:"local"`
	WebConfig struct {
		Switch     string `toml:"switch"`
		Ipaddr     string `toml:"ipaddr"`
		Port       string `toml:"port"`
		UserName   string `toml:"user_name"`
		UserPasswd string `toml:"user_passwd"`
	} `toml:"web_config"`
	Encryption struct {
		Switch      string `toml:"switch"`
		EncryptMode string `toml:"encrypt_mode"`
		Aes         string `toml:"aes"`
	} `toml:"encryption"`
	Rsa struct {
		PubKey string `toml:"pub_key"`
		PriKey string `toml:"pri_key"`
	} `toml:"rsa"`
}

type conPath struct {
	devPath string
	proPath string
	prePath string
}

// 遇到错误会panic
// recover捕获错误，并会在控制台输出log err以外的错误信息
// log err不会被recover恢复
func InitConfig() *AutoGenerated {
	// 读取配置文件
	var conf AutoGenerated
	c := conPath{
		devPath: filepath.FromSlash("/conf/dev/app.conf.toml"),
		proPath: filepath.FromSlash("/conf/pro/app.conf.toml"),
		prePath: filepath.FromSlash("/conf/pre/app.conf.toml"),
	}
	var path string
	pathHead, _ := os.Getwd()
	// 遍历结构体
	value := reflect.ValueOf(c)
	for i := 0; i < value.NumField(); i++ {
		_, err := os.Open(pathHead + value.Field(i).String())
		if err != nil {
			continue
		} else {
			sp := strings.SplitN(value.Field(i).String(), "/", 4)
			switch sp[2] {
			case "dev":
				ConfModel = ConfModelDev
				break
			case "pro":
				ConfModel = ConfModelPro
				break
			case "pre":
				ConfModel = ConfModelPre
				break
			}
		}
	}
	bind := map[ConfNum]string{ConfModelDev: c.devPath, ConfModelPro: c.proPath, ConfModelPre: c.prePath}
	for k, v := range bind {
		if k == ConfModel {
			path = pathHead + v
		}
	}
	if TestOnFilePath != "" {
		path = TestOnFilePath
	}

	defer utils.ReCoverErrorAndPrint()
	// 默认dev配置
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		log, err := logger.Std(nil)
		if err != this.Nil {
			panic(err)
		}
		defer log.Close()
		log.StdErrorLog("读取配置文件失败")
		panic(err)
	}
	return &conf
}

// 返回自定义错误,详见{common/error}
// 遇到错误返回，不在该层panic
func SaveTomlConfig(tomlFile *AutoGenerated) this.Error {
	path, _ := os.Getwd()
	file, err := os.Create(path + filepath.FromSlash("/conf/dev/app.conf.toml"))
	if err != nil {
		return this.SetError(err)
	}
	err = toml.NewEncoder(file).Encode(tomlFile)
	if err != nil {
		return this.Nil
	} else {
		return this.SetError(err)
	}
}
