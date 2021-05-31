package app

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	cf "github.com/mengzushan/bups/common/conf"
	"github.com/mengzushan/bups/common/encry"
	this "github.com/mengzushan/bups/common/error"
	"github.com/mengzushan/bups/common/logger"
	"github.com/mengzushan/bups/utils"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type InfoJson struct {
	CreateTime  int64  `json:"createTime"`
	CreateSize  int64  `json:"createSize"`
	UseTime     int64  `json:"useTime"`
	UseSize     int64  `json:"useSize"`
	WebPassword string `json:"webPassword"`
	DbPassword  string `json:"dbPassword"`
}

type ConfigJson struct {
	Rsa        string `json:"rsa"`
	Aes        string `json:"aes"`
	Key        string `json:"key"`
	Format     string `json:"format"`
	WebPath    string `json:"webPath"`
	StaticPath string `json:"staticPath"`
	LogPath    string `json:"logPath"`
	DbName     string `json:"dbName"`
}

func BackUpForFile() *ConfigJson {
	// 初始化日志
	std, _ := logger.Std(nil)
	defer std.Close()
	// 读取配置文件
	conf := cf.InitConfig()
	// 创建压缩包内的Json配置文件
	var enONToAes = "on"
	var enONToRsa = "on"
	// 密钥
	key := ""
	if conf.Encryption.Switch == "off" {
		enONToRsa = "off"
		enONToAes = "off"
		// 清空不需要的数据
		conf.Encryption.EncryptMode = ""
	}
	switch conf.Encryption.EncryptMode {
	case "aes":
		enONToRsa = "off"
		break
	case "rsa","rsa+aes","aes+rsa":
		var err error
		key, err = encryptKey()
		if err != this.Nil {
			return nil
		}
		break
	default:
		break
	}
	jsons := ConfigJson{
		Rsa:        enONToRsa,
		Aes:        enONToAes,
		Format:     "zip",
		Key:        key,
		WebPath:    conf.Local.Web,
		StaticPath: conf.Local.Static,
		LogPath:    conf.Local.Log,
		DbName:     conf.Database.DbName,
	}
	jsonf, _ := json.Marshal(&jsons)
	// 创建文件写入json
	pathHead, _ := os.Getwd()
	file, _ := os.Create(filepath.FromSlash(pathHead + "/cache/backup/config.json"))
	_, err := file.Write(jsonf)
	defer file.Close()
	if err != nil {
		std.StdErrorLog("Json配置文件写入失败")
		return nil
	}
	if conf.Local.Web != "" {
		//CreateZip(conf.Local.Web, "web.zip")
		err = Zip(conf.Local.Web, "web.zip")
	}
	if conf.Local.Static != "" {
		//CreateZip(conf.Local.Static, "static.zip")
		err = Zip(conf.Local.Static, "static.zip")
	}
	if conf.Local.Log != "" {
		err = Zip(conf.Local.Log, "log.zip")
	}
	if err != nil {
		std.StdErrorLog("文件压缩失败: " + err.Error())
		return nil
	} else {
		std.StdInfoLog(fmt.Sprintf("文件压缩成功: %s-%s-%s", conf.Local.Web, conf.Local.Log, conf.Local.Static))
		return &jsons
	}
}

func encryptKey() (string, this.Error) {
	key := ""
	var encrypt encry.CryptToRsa = &encry.CryptBlocks{}
	for i := 0; i < 16; i++ {
		rand.Seed(time.Now().UnixNano())
		k := rand.Intn(91-65) + 65
		key += string(byte(k))
	}
	// 读取公钥
	pathHead, _ := os.Getwd()
	pubPem, err := os.Open(pathHead + "/cache/rsa/public.pem")
	if err != nil {
		return "", this.SetError(err)
	}
	fileData, _ := ioutil.ReadAll(pubPem)
	keyByte, err := encrypt.EncryptToRsa([]byte(key), fileData)
	if err != this.Nil {
		return "", this.SetError(err)
	}
	// 使用base64编码加密之后的密钥
	return base64.StdEncoding.EncodeToString(keyByte), this.Nil
}

func CreateZip(srcPath string, createName string) {
	// 创建待写入的压缩文件
	//p,_ := os.Getwd()
	pathPrefix := "./cache/backup/"
	zipfile, err := os.Create(filepath.FromSlash(pathPrefix + createName))
	defer zipfile.Close()
	println(createName)
	if err != nil {
		std, _ := logger.Std(nil)
		std.StdErrorLog("文件创建失败" + filepath.FromSlash(pathPrefix+createName))
		panic(err)
	}
	// 创建压缩包流
	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	// 遍历目录
	filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取文件头信息
		header, _ := zip.FileInfoHeader(info)
		header.Name = strings.TrimPrefix(path, filepath.Dir(srcPath)+"/")
		// 判断文件是不是文件夹
		if info.IsDir() {
			header.Name += "/"
		} else {
			// 设置zip文件的压缩算法
			header.Method = zip.Deflate
		}
		// 创建压缩包头部信息
		w, _ := archive.CreateHeader(header)
		// 不是文件夹是将文件copy到流中
		if !info.IsDir() {
			file, _ := os.Open(path)
			defer file.Close()
			_, err = io.Copy(w, file)
			if err != nil {
				return err
			}
		}
		return err
	})

}

// srcFile could be a single file or a directory
func Zip(srcFile string, destZip string) error {
	// 判断destZip参数是不是传递路径
	var pwd string
	if utils.Equals(destZip, "/") {
		pwd = destZip
	} else {
		pwd = "./cache/backup/" + destZip
	}
	zipfile, err := os.Create(pwd)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		// header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return err
}

func ReadZipFile() {
	// 测试读取zip文件
	// Open a zip archive for reading.
	r, err := zip.OpenReader("/Users/harder/github.com-codes/bups/_build_0" + "/cache/backup/web.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		fmt.Printf("Contents of %s:\n", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.CopyN(os.Stdout, rc, 68)
		if err != nil {
			log.Fatal(err)
		}
		rc.Close()
		fmt.Println()
	}
}
