package cmd

import (
	"os"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/apiserver"
	"github.com/spf13/cobra"
)

var (
	apiServerPort    string
	apiServerRelease bool
)

// @title EagleEye API
// @version 0.2.0
// @description This is the API documentation for EagleEye.

// @host 127.0.0.1:9527
// @BasePath /api/v1
var apiserverCmd = cobra.Command{
	Use:   "apiserver",
	Short: "start api server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := apiserver.InitDB(); err != nil {
			apiserver.Logger.Error("apiserver.InitDB failed", "error", err)
			os.Exit(1)
		}
		defer apiserver.ReleaseDB()

		if err := apiserver.InitComponent(); err != nil {
			apiserver.Logger.Error("apiserver.InitComponent failed", "error", err)
			os.Exit(1)
		}
		defer apiserver.ReleaseComponent()

		router := apiserver.NewRouter()
		apiserver.RegisterRoutes(router, apiServerRelease)
		server := apiserver.NewServer(router, apiServerPort)

		if err := apiserver.Run(cmd.Context(), server); err != nil {
			apiserver.Logger.Error("apiserver.Run failed", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	apiserverCmd.Flags().StringVarP(&apiServerPort, "port", "p", "9527", "服务端口")
	apiserverCmd.Flags().BoolVarP(&apiServerRelease, "release", "r", false, "Release模式")
}
