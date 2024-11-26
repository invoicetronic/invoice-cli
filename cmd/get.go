/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/CIR2000/inv/models"
	"github.com/spf13/cobra"
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
	relativePath := "receive" + "/" + id
	fullURL := BuildUrl(relativePath)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	_, respBody := PerformRequest(req, &http.Client{})

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

	getCmd.Flags().BoolVar(&as_json, "json", false, "response as json")
	getCmd.Flags().StringVarP(&outdir, "dest", "d", "", "destination directory")
}
