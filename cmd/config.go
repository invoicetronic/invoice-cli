/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "configuration management",
	Long: `
Sets configuration options for the command.`,
	Run: func(cmd *cobra.Command, args []string) {
		list, _ := cmd.Flags().GetBool("list")
		if list {
			fmt.Println("Configuration file: " + viper.ConfigFileUsed())
			for key, value := range viper.AllSettings() {
				fmt.Printf("%s: %v\n", key, value)
			}
			os.Exit(0)
		}

		editor, _ := cmd.Flags().GetBool("edit")
		if editor {
			openEditor(viper.ConfigFileUsed())
			os.Exit(0)
		}
	},
}

var apikeyCmd = &cobra.Command{
	Use:   "apikey [value]",
	Short: "sets the API key (use with caution, it's sensitive data)",
	Long: `
Sets the API key. Use with caution, it's sensitive data.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value := args[0]
		viper.Set("apikey", value)
		if err := viper.WriteConfig(); err != nil {
			log.Fatalf("Error saving the configuration file: %v", err)
		}
		Verbose("api key set to: %s\n", value)
	},
}
var hostCmd = &cobra.Command{
	Use:   "host [value]",
	Short: "sets the remote host",
	Long: `
Sets the remote host. Use the full address, including the protocol.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value := args[0]
		viper.Set("host", value)
		if err := viper.WriteConfig(); err != nil {
			log.Fatalf("Error saving the configuration file: %v", err)
		}
		Verbose("host set to: %s\n", value)
	},
}

var verboseCmd = &cobra.Command{
	Use:   "verbose [bool]",
	Short: "sets the verbose mode",
	Long: `
Sets the verbose mode. Use 'true' or 'false'.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value, err := strconv.ParseBool(args[0])
		if err != nil {
			log.Fatal(err)
		}
		viper.Set("verbose", value)
		if err := viper.WriteConfig(); err != nil {
			log.Fatalf("Error saving the configuration file: %v", err)
		}
		Verbose("verbose set to: %v", value)
	},
}

func openEditor(filename string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-e", filename)
	case "windows":
		cmd = exec.Command("notepad", filename)
	default: // Linux e altri sistemi Unix-like
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano" // Fallback
		}
		cmd = exec.Command(editor, filename)
	}

	return cmd.Run()
}

func init() {
	configCmd.AddCommand(apikeyCmd, hostCmd, verboseCmd)
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().BoolP("list", "l", false, "list all")
	configCmd.Flags().BoolP("edit", "e", false, "open an editor")
}
