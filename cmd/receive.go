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

	"github.com/CIR2000/inv/models"
	"github.com/spf13/cobra"
)

var as_json bool
var unread bool
var outdir string
var remote_delete bool
var assume_yes bool

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
	receivePart := "receive"
	unreadPart := ""
	if unread {
		unreadPart = "/?unread=true"
	}
	relativePath := receivePart + unreadPart
	fullURL := BuildUrl(relativePath)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	_, respBody := PerformRequest(req, &http.Client{})

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
	} else {
		for _, item := range response.Items {
			ToFile(item.File_Name, item.Payload)
		}
	}
	if remote_delete && len(response.Items) > 0 {
		if !assume_yes {
			reader := bufio.NewReader(os.Stdin)

			fmt.Printf("Are you sure you want to remotely delete the documents? (y/N): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if strings.ToLower(input) == "n" || input == "" {
				fmt.Println("Remote delete canceleted.")
				return
			}
		}
		for _, item := range response.Items {
			remoteDelete(item.Id)
		}
	}

}
func remoteDelete(id int) {
	relativePath := ("receive/" + strconv.Itoa(id))
	fullURL := BuildUrl(relativePath)
	req, err := http.NewRequest("DELETE", fullURL, nil)
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

	receiveCmd.Flags().BoolVar(&as_json, "json", false, "response as json, no file will be saved")
	receiveCmd.Flags().BoolVarP(&assume_yes, "yes", "y", false, "assume yes on all answers")
	receiveCmd.Flags().BoolVar(&unread, "unread", false, "fetch unread documents only")
	receiveCmd.Flags().BoolVar(&remote_delete, "delete", false, "once the file has been downloaded, delete it from the remote API")
	receiveCmd.Flags().StringVarP(&outdir, "dest", "d", "", "destination directory")
}
