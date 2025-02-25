package cmd

import (
	"github.com/EscapeBearSecond/falcon/internal/meta"
	"github.com/spf13/cobra"
)

// metaCmd 输出程序元数据命令
var metaCmd = cobra.Command{
	Use:   "meta",
	Short: "print program metadata",
	Run: func(cmd *cobra.Command, args []string) {
		meta.Print()
	},
}
