package cmd

import (
	"bytes"
	"encoding/base64"
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

var signature string
var validate bool
var del bool

const auto = "auto"
const apply = "apply"
const none = "none"

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send invoice file(s)",
	Long: `
Send one or more invoice file(s) to the eInvoice API.

You can list multiple files and use wildcards. For example:

invoice send file1.xml file2.xml
invoice send dir/*.xml --delete`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		allowed := map[string]bool{
			auto:  true,
			apply: true,
			none:  true,
		}

		if !allowed[signature] {
			return fmt.Errorf("invalid signature: %s. Allowed values: %s, %s, %s", signature, auto, apply, none)
		}
		return nil
	},
	Run: sendRun,
}

func sendRun(_ *cobra.Command, args []string) {
	items := build(args)
	send(items)
}

func send(items []models.SendItem) {
	client := &http.Client{}

	url := BuildEndpointUrl("send")
	q := url.Query()
	q.Set("validate", strconv.FormatBool(validate))
	q.Set("signature", capitalizeFirst(signature))
	url.RawQuery = q.Encode()

	for _, item := range items {
		jsonBytes, _ := json.Marshal(item)

		req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(jsonBytes))
		if err != nil {
			log.Fatal(err)
		}

		resp, _ := PerformRequest(req, client)

		Verbose("%v sent (%v)", item.FileName, resp.Status)
		if del {
			err := os.Remove(item.FilePath)
			if err != nil {
				log.Fatalf("Error deleting %v: %v", item.FileName, err)
			}
			Verbose("%v deleted (--delete)", item.FileName)
		}
	}
}

func build(args []string) []models.SendItem {

	var items []models.SendItem
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
			item := models.SendItem{FilePath: file, FileName: filepath.Base(file), Payload: base64.StdEncoding.EncodeToString(content)}
			Verbose("%v selected and encoded (base64)", item.FileName)
			items = append(items, item)
		}
	}
	return items
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.Flags().BoolVar(&del, "delete", false, "once the file has been sent, delete it from disk")
	sendCmd.Flags().BoolVar(&validate, "validate", false, "validate first, and reject it the document is invalid")
	sendCmd.Flags().StringVar(&signature, "signature", auto, fmt.Sprintf("signature method (%s, %s, %s)", auto, apply, none))
}
