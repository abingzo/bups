package app

import (
	"github.com/mengzushan/bups/common/conf"
	"github.com/mengzushan/bups/common/encry"
	this "github.com/mengzushan/bups/common/error"
	"github.com/mengzushan/bups/common/logger"
	"github.com/mengzushan/bups/utils"
	"io/ioutil"
	"os"
	"encoding/base64"
	"path/filepath"
	"strings"
)

type options int

const (
	ENCRYPTOF  options = 0
	ENCRYPTON  options = 1
	FILEBUFFER string  = "/cache/backup"
)

// 参数为常量 ENCRYPTON或ENCRYPTOF,1开启加密,0则为关闭
func EncryptFile(eo options, cf *conf.AutoGenerated,backupConfigJson *ConfigJson,path string) this.Error {
	// 决定是否要备份
	// 在所有zip文件准备后完成

	// 初始化日志
	log, er := logger.Std(nil)
	if er != this.Nil {
		return er
	}
	defer log.Close()
	if eo == ENCRYPTOF {
		log.StdInfoLog("压缩文件不加密")
		return this.SetError("encrypt mode is off")
	}
	// 判断加密方式
	if backupConfigJson.Rsa == "off" && backupConfigJson.Aes == "on" {
		cf.Encryption.Aes = cf.Encryption.Aes
	} else if backupConfigJson.Rsa == "on" {
		// 使用私钥解密
		var encrypt encry.CryptToRsa = &encry.CryptBlocks{}
		pathHead, _ := os.Getwd()
		priPem,err := os.Open(pathHead + "/cache/rsa/private.pem")
		if err != nil {
			return this.SetError(err)
		}
		fileData, err := ioutil.ReadAll(priPem)
		if err != nil {
			return this.SetError(err)
		}
		// base64标准解码
		cipherText,err := base64.StdEncoding.DecodeString(backupConfigJson.Key)
		if err != nil {
			return this.SetError(err)
		}
		src, err := encrypt.DecryptToRsa(cipherText, fileData)
		if err != this.Nil {
			return this.SetError(err)
		}
		// 解密完成改变密钥
		cf.Encryption.Aes = string(src)
	} else {
		// 不加密
		return this.SetError("encrypt mode is off")
	}
	// 构造文件匹配列表
	fileList, err := MatchPathFile(cf)
	if err != this.Nil {
		log.StdErrorLog(err.Error())
		return this.SetError(err)
	}
	// 为列表中的每个文件添加路径
	for i := 0; i < len(fileList); i++ {
		fileList[i] = path + "/" + fileList[i]
	}
	_ = backToEncrypt(fileList,cf)
	return this.Nil
}

func MatchPathFile(cf *conf.AutoGenerated) ([]string, this.Error) {
	// 构造匹配列表
	pathList := matchToConf(cf)
	fileList := make([]string, 0)
	var flIndex uint
	pathf := func() string {
		pathHead, _ := os.Getwd()
		return pathHead + FILEBUFFER
	}()
	// 遍历缓存临时备份文件的目录
	err := filepath.Walk(pathf, func(path string, info os.FileInfo, err error) error {
		// 多次匹配以获得一个正确的结果
		var match bool
		for _, v := range pathList {
			if !match {
				match = utils.Equal(path, v)
			}
		}
		if err != nil {
			return err
		}
		// 匹配成功添加到文件列表
		if match {
			fileList = append(fileList, strings.TrimPrefix(path, pathf+"/"))
			if uint(len(pathList)-1) > flIndex {
				flIndex++
			}
		}
		return nil
	})
	if err != nil {
		return nil, this.SetError(err)
	}
	// utils封装的匹配两个字符串切片中的内容是否相同
	argMatchBool := utils.EqualToStrings(pathList, fileList)
	// 值相同且刨去空的配置项的情况下,值的个数相等即相同
	if !argMatchBool {
		return nil, this.SetError("arg num is not match")
	}
	return fileList, this.Nil
}

// 该层有panic
// 将备份的文件通过aes方式加密
func backToEncrypt(fl []string,config *conf.AutoGenerated) error {
	var en encry.Crypt = &encry.CryptBlocks{}
	log, err := logger.Std(nil)
	defer log.Close()
	// 捕获错误并打印信息
	defer utils.ReCoverErrorAndPrint()
	if err != this.Nil {
		panic(err)
	}
	// 循环打开文件
	for _, v := range fl {
		file, err := os.Open(v)
		if err != nil {
			log.StdErrorLog(err.Error())
			panic(err)
		}
		fileData, err := ioutil.ReadAll(file)
		if err != nil {
			log.StdErrorLog(err.Error())
			panic(err)
		}
		fileData, err = en.EncryptToAes(fileData, []byte(config.Encryption.Aes))
		if err != nil {
			log.StdErrorLog(err.Error())
			panic(err)
		}
		// 读取完毕关闭文件
		_ = file.Close()
		// 打开文件写入数据
		file, err = os.OpenFile(v, os.O_WRONLY, 0666)
		if err != nil {
			log.StdErrorLog(err.Error())
			panic(err)
		}
		_, err = file.Write(fileData)
		if err != nil {
			log.StdErrorLog(err.Error())
			panic(err)
		}
	}
	return nil
}

// 根据配置项匹配网站的文件夹地址，并指定返回值
func matchToConf(cf *conf.AutoGenerated) []string {
	pathList := make([]string, 0)
	if cf.Local.Web != "" {
		pathList = append(pathList, "web.zip")
	}
	if cf.Local.Static != "" {
		pathList = append(pathList, "static.zip")
	}
	if cf.Local.Log != "" {
		pathList = append(pathList, "log.zip")
	}
	if cf.Database.DbName != "" {
		pathList = append(pathList, cf.Database.DbName + ".zip")
	}
	return pathList
}
