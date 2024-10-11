package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/shirou/gopsutil/v3/net"

	"github.com/spf13/cobra"
)

// ifaceCmd 输出网卡信息命令
var ifaceCmd = cobra.Command{
	Use:   "iface",
	Short: "output net interfaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}

		f, err := os.Create("interfaces.csv")
		if err != nil {
			return fmt.Errorf("create interfaces out file failed: %w", err)
		}

		writer := csv.NewWriter(f)

		for i, interfaceInfo := range interfaces {
			writer.Write([]string{
				strconv.Itoa(i), interfaceInfo.Name,
			})
		}
		writer.Flush()

		return nil
	},
}
