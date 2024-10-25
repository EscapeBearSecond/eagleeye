package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/EscapeBearSecond/falcon/internal/license"
	"github.com/EscapeBearSecond/falcon/internal/util"
	"github.com/spf13/cobra"
)

var (
	licenseIdentity  string
	licenseAudience  string
	licenseExpiresAt string
	licenseHardwares []string

	licenseVerify string
	licenseSecret bool
)

var licenseCmd = cobra.Command{
	Use:   "license",
	Short: "generate license",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if licenseVerify != "" || licenseSecret {
			return nil
		}
		if licenseAudience == "" {
			return errors.New("audience is required")
		}
		if licenseExpiresAt == "" {
			return errors.New("expires_at is required")
		}
		if len(licenseHardwares) == 0 {
			return errors.New("hardwares is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if licenseSecret {
			secret := util.RandomStr(32)
			err := os.WriteFile("secret.txt", []byte(secret), 0644)
			if err != nil {
				return err
			}
			fmt.Println("secret:", secret)

			return nil
		}

		if licenseVerify != "" {
			l, err := license.Parse(licenseVerify)
			if err != nil {
				return err
			}

			if licenseIdentity != "" {
				license.UseSecret(licenseIdentity)
			}

			fmt.Println("license signature:", l.Signature)
			fmt.Println("calculate signature:", l.Sign())
			if err := l.Verify(); err != nil {
				fmt.Printf("verify failed: %s\n", err)
				return nil
			}
			fmt.Println("verify success")

			return nil
		}

		if licenseIdentity != "" {
			license.UseSecret(licenseIdentity)
		}
		err := license.Gen(licenseAudience, licenseExpiresAt, licenseHardwares...)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	licenseCmd.Flags().StringVarP(&licenseIdentity, "identity", "i", "", "license secret")
	licenseCmd.Flags().StringVarP(&licenseAudience, "audience", "a", "", "license audience")
	licenseCmd.Flags().StringVarP(&licenseExpiresAt, "expires_at", "e", "", "license expires at")
	licenseCmd.Flags().StringArrayVarP(&licenseHardwares, "hardwares", "w", nil, "license hardwares")

	licenseCmd.Flags().BoolVarP(&licenseSecret, "secret", "s", false, "generate secret")
	licenseCmd.Flags().StringVarP(&licenseVerify, "verify", "v", "", "verify license")
}
