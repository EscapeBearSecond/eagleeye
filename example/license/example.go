package main

import (
	"log"
	"os"

	"github.com/EscapeBearSecond/eagleeye/pkg/license"
)

func main() {
	stdlog := log.New(os.Stderr, "", log.LstdFlags)

	// 使用自定义secret
	// license.UseSecret("my-secret")

	// 未watch证书，验证失败
	err := license.Verify()
	if err != nil {
		stdlog.Printf("verify failed: %s\n", err)
	}

	// 监听证书文件
	watcher, err := license.Watch("./license.json")
	if err != nil {
		panic(err)
	}
	// 停止监听
	defer watcher.Stop()

	// 验证证书
	if err := license.Verify(); err != nil {
		stdlog.Printf("verify failed: %s\n", license.Verify())
	} else {
		stdlog.Printf("verify success\n")
	}

	// 获取当前证书信息
	stdlog.Printf("license: %+v", license.L())
}
