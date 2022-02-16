package common

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/wonderivan/logger"
)

func init() {
	var confile string
	flag.StringVar(&confile, "f", "", "配置文件路径")
	flag.Parse()

	_, err := os.Stat(confile)
	if err != nil {
		flag.Usage()
		os.Exit(0)
	}
	args := strings.Split(os.Args[0], "/")
	pgnm := args[len(args)-1]
	path, err := os.Getwd()
	if err != nil {
		log.Fatal("获取运行目录失败")
	}
	logconf = strings.ReplaceAll(logconf, "#entity", pgnm)
	logger.SetLogger(logconf)
	logger.SetLogPathTrim(path)
	viper.SetConfigFile(confile)
	if e := viper.ReadInConfig(); e != nil {
		log.Panicln("配置文件读取失败", e)
	}
}

var logconf string = `{"TimeFormat":"2006-01-02 15:04:05","File": {"filename": "/tmp/logs/#entity.app.log","level": "TRAC","daily": true,"maxlines": 1000000,"maxsize": 1,"maxdays": -1,"append": true,"permit": "0660"}}`
