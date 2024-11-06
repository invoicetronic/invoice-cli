/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var apiKey string
var host string
var version int
var verbose bool

const product_name string = "Fatture API"
const default_host string = "https://localhost:7019/api/"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "inv",
	Short: "Send and receive invoices via " + product_name,
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func BasicAuth() string {
	auth := viper.GetString("apikey") + ":" + ""
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func PerformRequest(req *http.Request, client *http.Client) (*http.Response, []byte) {
	req.Header.Set("Authorization", "Basic "+BasicAuth())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		log.Printf("Send failed (%v)", resp.Status)
		if len(respBody) > 0 {
			log.Println(string(respBody))
		}
		os.Exit(1)
	}

	return resp, respBody
}

func init() {

	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config_file", "", "config file (default is $HOME/.inv.yaml)")

	rootCmd.PersistentFlags().StringVar(&apiKey, "apikey", "", "your API key")
	viper.BindPFlag("apikey", rootCmd.PersistentFlags().Lookup("apikey"))

	rootCmd.PersistentFlags().StringVar(&host, "host", default_host, "API base address")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))

	rootCmd.PersistentFlags().IntVar(&version, "version", 1, "API version")
	viper.BindPFlag("version", rootCmd.PersistentFlags().Lookup("version"))

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "display more verbose outut in console output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

}

func ToFile(filename string, payload string) {
	if verbose {
		log.Printf("Decoding payload for %v\n", filename)
	}
	decodedData, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		log.Fatalf("Error decoding "+filename+": %v", err)
	}

	var filePath string
	if outdir != "" {
		filePath, err = getFullFilePath(outdir, filename)
		if err != nil {
			log.Fatal(err)
		}

		err = createDirectoryIfNotExists(outdir)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		filePath = filename
	}

	if verbose {
		log.Printf("Creating file %v\n", filename)
	}
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error creating "+filename+": %v", err)
	}
	defer file.Close()

	if verbose {
		log.Printf("Writing to file %v\n", filename)
	}
	_, err = file.Write(decodedData)
	if err != nil {
		log.Fatalf("Error writing to file "+filename+": %v", err)
	}

	if verbose {
		log.Printf("Write to %v succeded\n", filename)
	}
}

func createDirectoryIfNotExists(dest string) error {
	// Creare la directory se non esiste
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.SetDefault("host", default_host)
	viper.SetDefault("version", 1)
	viper.SetDefault("verbose", false)
	viper.SetDefault("apikey", "")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".inv")
		viper.AddConfigPath(home)

		viper.SafeWriteConfigAs(filepath.Join(home, ".inv.yaml"))
	}
	viper.SetEnvPrefix("inv")
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
