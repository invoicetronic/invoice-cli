package cmd

import (
	"encoding/base64"
	"fmt"
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

const version = "1.0.0"
const productName string = "Invoicetronic API"
const defaultHost string = "https://api.invoicetronic.com"
const apiVersion int = 1
const userAgent string = "Invoicetronic-CLI/" + version

var cfgFile string
var apiKey string
var host string
var verbose bool
var home, _ = os.UserHomeDir()
var defaultConfigFile = filepath.Join(home, ".invoice.yaml")

var rootCmd = &cobra.Command{
	Use:   "invoice",
	Short: "send and receive invoice file(s) via the " + productName,
	Long: `
Invoice is a CLI command to exchange electronic invoices with the Servizio di 
Interscambio (SDI), the official Italian invoice exchange service. 

It leverages Invoicetronic API to quickly and seamlessly send and receive 
invoices from the command line.

For more information, please visit https://invoicetronic.com`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Version: version,
}

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

func performRequest(req *http.Request, client *http.Client) (*http.Response, []byte, error) {
	req.Header.Set("Authorization", "Basic "+BasicAuth())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return resp, respBody, fmt.Errorf("request failed (%v): %s", resp.Status, string(respBody))
	}

	return resp, respBody, nil
}

func PerformRequest(req *http.Request, client *http.Client) (*http.Response, []byte) {
	resp, body, err := performRequest(req, client)
	if err != nil {
		if resp != nil {
			log.Printf("Send failed (%v)", resp.Status)
			if len(body) > 0 {
				log.Println(string(body))
			}
			os.Exit(1)
		}
		log.Fatal(err)
	}
	return resp, body
}

func init() {

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config_file", defaultConfigFile, "configuration file")

	rootCmd.PersistentFlags().StringVar(&apiKey, "apikey", "", "api key for the "+productName)
	_ = viper.BindPFlag("apikey", rootCmd.PersistentFlags().Lookup("apikey"))

	rootCmd.PersistentFlags().StringVar(&host, "host", defaultHost, "host address")
	_ = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "display a more verbose output")
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

}

func toFile(filename, payload, encoding, destDir string) error {
	Verbose("Processing payload for %v with encoding %v\n", filename, encoding)
	var decodedData []byte

	if encoding == "Base64" {
		Verbose("Decoding base64 payload for %v\n", filename)
		var err error
		decodedData, err = base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return fmt.Errorf("error decoding base64 payload for %v: %v", filename, err)
		}
	} else {
		Verbose("Using payload as plain string for %v\n", filename)
		decodedData = []byte(payload)
	}

	var filePath string
	if destDir != "" {
		var err error
		filePath, err = getFullFilePath(destDir, filename)
		if err != nil {
			return err
		}
		err = createDirectoryIfNotExists(destDir)
		if err != nil {
			return err
		}
	} else {
		filePath = filename
	}

	Verbose("Creating file %v\n", filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating %s: %v", filename, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	Verbose("Writing to file %v\n", filename)
	_, err = file.Write(decodedData)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %v", filename, err)
	}

	Verbose("Write to %v succeeded\n", filename)
	return nil
}

func ToFile(filename, payload, encoding string) {
	if err := toFile(filename, payload, encoding, outputDir); err != nil {
		log.Fatal(err)
	}
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

func Verbose(format string, v ...any) {
	if !verbose {
		return
	}
	log.Printf(format, v...)
}

func BuildEndpointUrl(elem ...string) *url.URL {
	endpointUrl, _ := url.Parse(viper.GetString("host"))
	joinedPath := append([]string{endpointUrl.Path, "/v" + strconv.Itoa(apiVersion)}, elem...)
	endpointUrl.Path = path.Join(joinedPath...)
	return endpointUrl
}

func initConfig() {

	viper.SetDefault("host", defaultHost)
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

		_ = viper.SafeWriteConfigAs(defaultConfigFile)
	}
	viper.SetEnvPrefix("invoice")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}
