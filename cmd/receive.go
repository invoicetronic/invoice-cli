/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/invoicetronic/invoice/models"
	"github.com/spf13/cobra"
)

var as_json bool
var unread bool
var outdir string
var remote_delete bool
var assume_yes bool

var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "receive invoice file(s)",
	Long: `
Download one or more invoice file(s) from the API.`,
	Run: receiveRun,
}

func receiveRun(cmd *cobra.Command, args []string) {
	url := BuildEndpointUrl("receive")
	q := url.Query()
	q.Set("unread", strconv.FormatBool(unread))
	url.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	_, respBody := PerformRequest(req, &http.Client{})

	var items []models.ReceiveItem
	if err := json.Unmarshal(respBody, &items); err != nil {
		log.Fatal(err)
	}

	if as_json {
		jsonData, err := json.Marshal(items)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(jsonData))
	} else {
		for _, item := range items {
			ToFile(item.File_Name, item.Payload)
		}
	}
	if remote_delete && len(items) > 0 {
		if !assume_yes {
			reader := bufio.NewReader(os.Stdin)

			fmt.Printf("Are you sure you want to remotely delete the documents? (y/N): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if strings.ToLower(input) == "n" || input == "" {
				fmt.Println("Remote delete canceled.")
				return
			}
		}
		for _, item := range items {
			remoteDelete(item.Id)
		}
	}

}
func remoteDelete(id int) {
	url := BuildEndpointUrl("receive", strconv.Itoa(id))
	req, err := http.NewRequest("DELETE", url.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	PerformRequest(req, &http.Client{})
}

func getFullFilePath(dest, filename string) (string, error) {
	if !strings.HasSuffix(dest, string(os.PathSeparator)) {
		dest += string(os.PathSeparator)
	}

	return filepath.Join(dest, filename), nil
}

func init() {
	rootCmd.AddCommand(receiveCmd)

	receiveCmd.Flags().BoolVar(&as_json, "json", false, "response as json; no file will be saved")
	receiveCmd.Flags().BoolVarP(&assume_yes, "yes", "y", false, "assume yes on all answers")
	receiveCmd.Flags().BoolVarP(&unread, "unread", "u", false, "fetch unread documents only")
	receiveCmd.Flags().BoolVar(&remote_delete, "delete", false, "once the file has been downloaded, delete it from the API")
	receiveCmd.Flags().StringVarP(&outdir, "dest", "d", "", "destination directory")
}
