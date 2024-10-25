package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/EscapeBearSecond/falcon/internal/export"
	"github.com/EscapeBearSecond/falcon/pkg/report"
	"github.com/EscapeBearSecond/falcon/pkg/types"
	"github.com/EscapeBearSecond/falcon/pkg/util"
	"github.com/rs/xid"
	"github.com/spf13/cobra"
)

var (
	hostDiscoveryFile string
	portScanningFile  string
	vulnJobFile       string
)

var reportCmd = cobra.Command{
	Use:   "report",
	Short: "generate report",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if hostDiscoveryFile == "" && portScanningFile == "" && vulnJobFile == "" {
			return errors.New("no result files specified")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		resultReaders := make([]*types.ResultReader, 0)

		if hostDiscoveryFile != "" {
			hd, err := os.Open(hostDiscoveryFile)
			if err != nil {
				return fmt.Errorf("open host discovery result file failed: %w", err)
			}
			defer hd.Close()

			resultReaders = append(resultReaders, &types.ResultReader{
				Format: export.Format(hostDiscoveryFile),
				Stage:  types.StageHostDiscovery,
				Reader: hd,
			})
		}

		if portScanningFile != "" {
			ps, err := os.Open(portScanningFile)
			if err != nil {
				return fmt.Errorf("open port scanning result file failed: %w", err)
			}
			defer ps.Close()

			resultReaders = append(resultReaders, &types.ResultReader{
				Format: export.Format(portScanningFile),
				Stage:  types.StagePortScanning,
				Reader: ps,
			})
		}

		if vulnJobFile != "" {
			f, err := os.Open(vulnJobFile)
			if err != nil {
				return fmt.Errorf("open job result file failed: %w", err)
			}
			defer f.Close()

			resultReaders = append(resultReaders, &types.ResultReader{
				Format: export.Format(vulnJobFile),
				Stage:  types.StageJob,
				Reader: f,
			})
		}

		result, err := util.ReloadResult(resultReaders...)
		if err != nil {
			return err
		}

		result.EntryID = xid.New().String()
		err = report.Generate(
			report.WithDirectory("."),
			report.WithJobIndexes(0),
			report.WithEntryResult(result),
		)

		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	reportCmd.Flags().StringVarP(&hostDiscoveryFile, "host_discovery", "d", "", "在线扫描结果文件")
	reportCmd.Flags().StringVarP(&portScanningFile, "port_scanning", "p", "", "端口扫描结果文件")
	reportCmd.Flags().StringVarP(&vulnJobFile, "job", "j", "", "漏洞扫描任务结果文件")
}
