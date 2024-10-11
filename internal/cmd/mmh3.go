package cmd

import (
	"errors"
	"fmt"

	"github.com/EscapeBearSecond/eagleeye/internal/mmh3"
	"github.com/spf13/cobra"
)

var (
	mmh3Base64   string
	mmh3Base64Py string
)

var mmh3Cmd = cobra.Command{
	Use:   "mmh3",
	Short: "mmh3 hash",
	RunE: func(cmd *cobra.Command, args []string) error {
		if mmh3Base64 == "" && mmh3Base64Py == "" {
			return errors.New("mmh3Base64 or mmh3Base64Py is required")
		}

		if mmh3Base64 != "" && mmh3Base64Py != "" {
			return errors.New("mmh3Base64 and mmh3Base64Py are mutually exclusive")
		}

		var input string
		var base64Func mmh3.Base64Func
		if mmh3Base64 != "" {
			input = mmh3Base64
			base64Func = mmh3.Base64Encode
		} else if mmh3Base64Py != "" {
			input = mmh3Base64Py
			base64Func = mmh3.Base64PyEncode
		}

		hash, err := mmh3.Hash(input, base64Func)
		if err != nil {
			return fmt.Errorf("mmh3 hash failed: %w", err)
		}

		fmt.Println(hash)

		return nil
	},
}

func init() {
	mmh3Cmd.Flags().StringVar(&mmh3Base64, "base64", "", "调用base64编码")
	mmh3Cmd.Flags().StringVar(&mmh3Base64Py, "base64_py", "", "调用base64_py编码")
}
