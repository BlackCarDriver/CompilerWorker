package config

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"io/ioutil"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
)

type serverConfig struct {
	MasterURI  string `xml:"master_uri"`  // 请求注册服务的uri
	ServerAddr string `xml:"server_addr"` // 本地监听的地址

	S2SName    string `xml:"s2s_name"`    // 服务名称
	S2SKey     string `xml:"s2s_key"`     // 验证节点的依据
	Secret     string `xml:"secret"`      // 知道secret才能计算得到正确的s2skey
	Tag        string `xml:"tag"`         // 节点标记
	RequestURI string `xml:"request_uri"` // 客户端请求本节点提供服务的URI

	LogPath string `xml:"log_path"` // 日志存储的位置(斜杠结尾)
	IsTest  bool   `xml:"is_test"`
}

var ServerConfig serverConfig

func init() {
	xmlFile, err := os.Open("./config/config.xml")
	if err != nil {
		logs.Critical("Error opening config file: %v", err)
		os.Exit(1)
		return
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)
	xml.Unmarshal(b, &ServerConfig)

	// 一些检查和修正
	ServerConfig.LogPath = strings.TrimRight(ServerConfig.LogPath, "/") + "/"

	// 计算s2sKey
	if ServerConfig.S2SKey == "" {
		md5Ctx := md5.New()
		md5Ctx.Write([]byte(ServerConfig.Secret + ServerConfig.S2SName + ServerConfig.RequestURI))
		ServerConfig.S2SKey = hex.EncodeToString(md5Ctx.Sum(nil))
	}

	logs.Info("ServerConfig: %+v", ServerConfig)
}
