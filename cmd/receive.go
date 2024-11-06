/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CIR2000/inv/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var as_json bool
var outdir string

var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive invoices",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: receiveRun,
}

func receiveRun(cmd *cobra.Command, args []string) {
	baseURL, _ := url.Parse(viper.GetString("host") + "v" + strconv.Itoa(viper.GetInt("version")) + "/")
	relativePath, _ := url.Parse("receive")
	fullURL := baseURL.ResolveReference(relativePath).String()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Fatalf("Error creating a receive request: %v", err)
	}

	req.Header.Set("Authorization", "Basic "+MyBasicAuth())
	req.Header.Set("Accept", "application/json")

	if verbose {
		log.Println("Requesting items...")
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if !(isSuccessStatusCode(resp.StatusCode)) {
		log.Printf("Receive failed (%v)", resp.Status)
		if len(respBody) > 0 {
			log.Println(string(respBody))
		}
		os.Exit(1)
	}

	var response models.Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Fatal(err)
	}

	if as_json {
		jsonData, err := json.Marshal(response.Items)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(jsonData))
	}

	for _, item := range response.Items {
		ToFile(item.File_Name, item.Payload)
	}

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

func getFullFilePath(dest, filename string) (string, error) {
	// Assicurarsi che la directory di destinazione abbia il separatore finale
	if !strings.HasSuffix(dest, string(os.PathSeparator)) {
		dest += string(os.PathSeparator)
	}

	// Costruire il filepath completo
	return filepath.Join(dest, filename), nil
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
func init() {
	rootCmd.AddCommand(receiveCmd)

	receiveCmd.Flags().BoolVar(&as_json, "json", false, "response as json")
	receiveCmd.Flags().StringVarP(&outdir, "dest", "d", "", "destination directory")
}
