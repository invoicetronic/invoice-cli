/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/CIR2000/inv/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:   "get [integer]",
	Short: "Get an invoice by ID",
	Args:  cobra.ExactArgs(1),
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: getRun,
}

func getRun(cmd *cobra.Command, args []string) {
	id := args[0]
	baseURL, _ := url.Parse(viper.GetString("host") + "v" + strconv.Itoa(viper.GetInt("version")) + "/")
	relativePath, _ := url.Parse("receive" + "/" + id)
	fullURL := baseURL.ResolveReference(relativePath).String()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Fatalf("Error creating a receive request: %v", err)
	}

	req.Header.Set("Authorization", "Basic "+MyBasicAuth())
	req.Header.Set("Accept", "application/json")

	if verbose {
		log.Printf("Requesting item #%v\n", id)
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
		log.Printf("Get failed (%v)", resp.Status)
		if len(respBody) > 0 {
			log.Println(string(respBody))
		}
		os.Exit(1)
	}

	var response models.ReceiveItem
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Fatal(err)
	}

	if as_json {
		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(jsonData))
	}

	ToFile(response.File_Name, response.Payload)

}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolVar(&as_json, "json", false, "Response as json")
	getCmd.Flags().StringVarP(&outdir, "dest", "d", "", "Destination directory")
}
