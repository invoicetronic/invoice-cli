package models

type SendItem struct {
	Payload   string
	File_Name string
	FilePath  string
}

type ReceiveItem struct {
	Id        int
	Payload   string `json:"payload"`
	File_Name string `json:"file_name"`
}
type Response struct {
	Items []ReceiveItem `json:"$values"`
}
