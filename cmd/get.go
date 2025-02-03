/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/invoicetronic/invoice/models"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [integer]",
	Short: "get an invoice by ID",
	Long: `
Get an invoice by ID.`,
	Args: cobra.ExactArgs(1),
	Run:  getRun,
}

func getRun(cmd *cobra.Command, args []string) {
	id := args[0]
	url := BuildEndpointUrl("receive", id)
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	_, respBody := PerformRequest(req, &http.Client{})

	var response models.ReceiveItem
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Fatal(err)
	}

	if asJson {
		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(jsonData))
	}

	ToFile(response.FileName, response.Payload)

}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolVar(&asJson, "json", false, "response as json")
	getCmd.Flags().StringVarP(&outputDir, "dest", "d", "", "destination directory")
}
