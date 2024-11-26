/*
Copyright © 2024 Nicola Iarocci & CIR 2000
*/
package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/CIR2000/inv/models"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send one or more invoices",
	Long: `Send one or more invoice file(s) to the API.
You can list multiple files and use wildcards. For example:

invoice send file1.xml file2.xml
invoice send dir/*.xml --delete`,
	Run: sendRun,
}

func sendRun(cmd *cobra.Command, args []string) {
	items := build(args)
	send(cmd, items)
}

func send(cmd *cobra.Command, items []models.SendItem) {
	client := &http.Client{}
	validate, _ := cmd.Flags().GetBool("validate")
	sendPart := "send"
	validatePart := ""
	if validate {
		validatePart = "/?validate=true"
	}
	relativePath := (sendPart + validatePart)
	fullURL := BuildUrl(relativePath)

	for _, item := range items {
		json, _ := json.Marshal(item)
		jsonBytes := []byte(json)

		req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonBytes))
		if err != nil {
			log.Fatal(err)
		}

		resp, _ := PerformRequest(req, client)

		toVerbose("%v sent successfully (%v)", item.File_Name, resp.Status)
		delete, _ := cmd.Flags().GetBool("delete")
		if delete {
			err := os.Remove(item.FilePath)
			if err != nil {
				log.Fatalf("Error deleting %v: %v", item.File_Name, err)
			}
			toVerbose("%v deleted (--delete)", item.File_Name)
		}
	}

}

func build(args []string) []models.SendItem {

	items := []models.SendItem{}
	for _, arg := range args {
		files, err := filepath.Glob(arg)
		if err != nil {
			log.Fatalf("Error parsing the file names: %v", err)
		}

		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}
			item := models.SendItem{FilePath: file, File_Name: filepath.Base(file), Payload: base64.StdEncoding.EncodeToString(content)}
			toVerbose("%v selected and encoded (base64)", item.File_Name)
			items = append(items, item)
		}
	}
	return items
}

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.Flags().Bool("delete", false, "once the file has been sent, delete it from disk")
	sendCmd.Flags().Bool("validate", false, "validate first, and reject it the document is invalid")
}
