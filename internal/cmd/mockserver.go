package cmd

import (
	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/mockserver"
	"github.com/spf13/cobra"
)

var (
	mockServerPort string
)

// mockServerCmd 测试服务器命令
var mockServerCmd = cobra.Command{
	Use:   "mockserver",
	Short: "start a http mock server",
	Run: func(cmd *cobra.Command, args []string) {
		mockserver.Serve(cmd.Context(), mockServerPort)
	},
}

func init() {
	mockServerCmd.Flags().StringVarP(&mockServerPort, "port", "p", "9080", "测试服务端口")
}
