package cmd

import (
	"fmt"
	"os"

	utils "github.com/cenk1cenk2/do-dyndns/utils"
	"github.com/spf13/cobra"
)

// Version get current version of application
var Version string = "__VERSION__"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "do-dyndns",
	Short:   "Dynamically set your subdomains IP addresses that utilize Digital Ocean nameservers.",
	Version: Version,
	Run:     func(cmd *cobra.Command, args []string) { run(cmd, args) },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		utils.Log.Fatal(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	utils.Log.Println("test")

	utils.Log.Println(utils.Config.Domains)
	utils.Log.Println(utils.Config.Subdomains)
	fmt.Println(utils.Config)
}

func init() {
	fmt.Println("|d|o|-|d|y|n|d|n|s|", fmt.Sprintf("v%s", Version))

	// initialize
	cobra.OnInitialize(utils.InitConfig)

	// persistent flags
	rootCmd.PersistentFlags().StringVar(&utils.Cfg, "config", "", "config file ({.,/etc/do-dyndns/,$HOME}/.do-dyndns.yml)")
}
