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

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		list, _ := cmd.Flags().GetBool("list")
		if list {
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
	Short: "sets the API key and saves it for reuse (use with caution)",
	Args:  cobra.ExactArgs(1), // Richiede esattamente un argomento
	Run: func(cmd *cobra.Command, args []string) {
		value := args[0]
		viper.Set("apikey", value)
		if err := viper.WriteConfig(); err != nil {
			log.Fatalf("Error saving the configuration file: %v", err)
		}
		if verbose {
			fmt.Printf("API key set to: %s\n", value)
		}
	},
}
var hostCmd = &cobra.Command{
	Use:   "host [value]",
	Short: "sets the remote host (https://example.com/) and saves it",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value := args[0]
		viper.Set("host", value)
		if err := viper.WriteConfig(); err != nil {
			log.Fatalf("Error saving the configuration file: %v", err)
		}
		if verbose {
			fmt.Printf("host set to: %s\n", value)
		}
	},
}

var verboseCmd = &cobra.Command{
	Use:   "verbose [bool]",
	Short: "sets the verbose mode and saves it",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value, err := strconv.ParseBool(args[0])
		if err != nil {
			log.Fatal(err)
		}
		viper.Set("verbose", value)
		if err := viper.WriteConfig(); err != nil {
			log.Fatalf("Error saving the configuration file: %v", err)
		}
		if verbose {
			fmt.Printf("verbose set to: %v", value)
		}
	},
}
var versionCmd = &cobra.Command{
	Use:   "version [value]",
	Short: "sets the version of the API",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		value, err := strconv.ParseInt(args[0], 0, 16)
		if err != nil {
			log.Fatal(err)
		}
		viper.Set("version", value)
		if err := viper.WriteConfig(); err != nil {
			log.Fatalf("Error saving the configuration file: %v", err)
		}
		if verbose {
			fmt.Printf("verbose set to: %v", value)
		}
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
	configCmd.AddCommand(apikeyCmd, hostCmd, verboseCmd, versionCmd)
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().BoolP("list", "l", false, "list all")
	configCmd.Flags().BoolP("edit", "e", false, "open an editor")
}
