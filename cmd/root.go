/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)


const product_name string = "eInvoice API"
const default_host string = "https://api.invoicetronic.com"
const api_version int = 1

var cfgFile string
var apiKey string
var host string
var verbose bool
var home, _ = os.UserHomeDir()
var default_config_file string = filepath.Join(home,".invoice.yaml")

var rootCmd = &cobra.Command{
	Use:   "invoice",
	Short: "send and receive invoice file(s) via " + product_name,
	Long: `
Invoice is a CLI command to exchange electronic invoices with the Servizio di 
Interscambio (SDI), the official Italian invoice exchange service. 

It leverages Invoicetronic's eInvoice API to quickly and seamlessly send and 
receive invoices from the command line.

For more information, please visit https://invoicetronic.com.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config_file", default_config_file, "configuration file")

	rootCmd.PersistentFlags().StringVar(&apiKey, "apikey", "", "api key for "+product_name)
	viper.BindPFlag("apikey", rootCmd.PersistentFlags().Lookup("apikey"))

	rootCmd.PersistentFlags().StringVar(&host, "host", default_host, "host address")
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "display a more verbose output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

}

func ToFile(filename string, payload string) {
	toVerbose("Decoding payload for %v\n", filename)
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

	toVerbose("Creating file %v\n", filename)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error creating "+filename+": %v", err)
	}
	defer file.Close()

	toVerbose("Writing to file %v\n", filename)
	_, err = file.Write(decodedData)
	if err != nil {
		log.Fatalf("Error writing to file "+filename+": %v", err)
	}

	toVerbose("Write to %v succeded\n", filename)
}

func createDirectoryIfNotExists(dest string) error {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func toVerbose(format string, v ...any) {
	if !verbose {
		return
	}
	log.Printf(format, v...)
}

func BuildUrl(relativePath string) string {
	base, _ := url.Parse(viper.GetString("host"))
	base.Path = path.Join(base.Path, "/v"+strconv.Itoa(api_version), relativePath)
	return base.String()
}

func initConfig() {

	viper.SetDefault("host", default_host)
	viper.SetDefault("verbose", false)
	viper.SetDefault("apikey", "")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".invoice")
		viper.AddConfigPath(home)

		viper.SafeWriteConfigAs(default_config_file)
	}
	viper.SetEnvPrefix("invoice")
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
