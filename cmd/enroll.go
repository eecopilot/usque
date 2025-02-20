package cmd

import (
	"crypto/x509"
	"log"

	"github.com/Diniboy1123/usque/api"
	"github.com/Diniboy1123/usque/config"
	"github.com/Diniboy1123/usque/models"
	"github.com/spf13/cobra"
)

var enrollCmd = &cobra.Command{
	Use:   "enroll",
	Short: "Enrolls a MASQUE private key and switches mode",
	Long: "Enrolls a MASQUE private key and switches mode. Useful for ZeroTier where IPv6 address can change." +
		" Or if you just want to deploy a new key.",
	Run: func(cmd *cobra.Command, args []string) {
		if !config.ConfigLoaded {
			cmd.Println("Config not loaded. Please register first.")
			return
		}

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("Failed to get config path: %v", err)
		}
		if configPath == "" {
			log.Fatalf("Config path is required")
		}

		deviceName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("Failed to get device name: %v", err)
		}

		log.Printf("Enrolling device key...")

		accountData := models.AccountData{
			Token: config.AppConfig.AccessToken,
			ID:    config.AppConfig.ID,
		}

		privKey, err := config.AppConfig.GetEcPrivateKey()
		if err != nil {
			log.Fatalf("Failed to get private key: %v", err)
		}

		publicKey, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
		if err != nil {
			log.Fatalf("Failed to marshal public key: %v", err)
		}

		updatedAccountData, err := api.EnrollKey(accountData, publicKey, deviceName)
		if err != nil {
			log.Fatalf("Failed to enroll device key: %v", err)
		}

		log.Printf("Successful registration. Saving config...")

		config.AppConfig = config.Config{
			PrivateKey: config.AppConfig.PrivateKey,
			// TODO: proper endpoint parsing in utils
			// strip :0
			EndpointV4: updatedAccountData.Config.Peers[0].Endpoint.V4[:len(updatedAccountData.Config.Peers[0].Endpoint.V4)-2],
			// strip [ from beginning and ]:0 from end
			EndpointV6:     updatedAccountData.Config.Peers[0].Endpoint.V6[1 : len(updatedAccountData.Config.Peers[0].Endpoint.V6)-4],
			EndpointPubKey: updatedAccountData.Config.Peers[0].PublicKey,
			License:        updatedAccountData.Account.License,
			ID:             updatedAccountData.ID,
			AccessToken:    accountData.Token,
			IPv4:           updatedAccountData.Config.Interface.Addresses.V4,
			IPv6:           updatedAccountData.Config.Interface.Addresses.V6,
		}

		config.AppConfig.SaveConfig(configPath)

		log.Printf("Config saved to %s", configPath)
	},
}

func init() {
	enrollCmd.Flags().StringP("name", "n", "", "Rename device a given name")
	rootCmd.AddCommand(enrollCmd)
}
