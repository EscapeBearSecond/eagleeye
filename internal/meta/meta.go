package meta

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

var (
	BuildTime    string = time.Now().Format("2006-01-02 15:04:05") //构建时间
	BuildUser    string = "Unknown"                                //构建用户
	BuildHost, _        = os.Hostname()                            //构建主机
	BuildVer     string = "0.0.0-dev"                              //构建版本
	BuildCommit  string = "Unknown"
	BuildBranch  string = "Unknown"

	BuildOS   = runtime.GOOS
	BuildArch = runtime.GOARCH
)

// Print 打印程序元数据
func Print() {
	fmt.Println("BuildGoVer:", runtime.Version())
	fmt.Println("BuildOS:", BuildOS)
	fmt.Println("BuildArch:", BuildArch)
	fmt.Println("BuildUser:", BuildUser)
	fmt.Println("BuildHost:", BuildHost)
	fmt.Println("BuildTime:", BuildTime)
	fmt.Println("BuildVer:", "v"+BuildVer)
	fmt.Println("BuildBranch:", BuildBranch)
	fmt.Println("BuildCommit:", BuildCommit)
}
