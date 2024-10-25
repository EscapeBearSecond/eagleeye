package cmd

import (
	"errors"

	"github.com/EscapeBearSecond/falcon/internal/tvalidator"
	"github.com/spf13/cobra"
)

var (
	fs bool

	driver string
	dsn    string
	sql    string

	dir string
)

var templateCmd = cobra.Command{
	Use:   "template",
	Short: "template",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if fs {
			if dir == "" {
				return errors.New("fs mode requires dir")
			}
		} else {
			if driver == "" || dsn == "" || sql == "" {
				return errors.New("db mode requires driver, dsn and sql")
			}
			if driver != "mysql" && driver != "postgres" {
				return errors.New("invalid driver, only support mysql and postgres")
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if fs {
			return tvalidator.ExecuteFsCommand(dir)
		} else {
			return tvalidator.ExecuteDbCommand(driver, dsn, sql)
		}
	},
}

func init() {
	templateCmd.Flags().BoolVar(&fs, "fs", false, "db")

	templateCmd.Flags().StringVar(&driver, "driver", "", "database driver")
	templateCmd.Flags().StringVar(&dsn, "dsn", "", "database dsn")
	templateCmd.Flags().StringVar(&sql, "sql", "", "query sql")

	templateCmd.Flags().StringVar(&dir, "dir", "", "directory")
}
