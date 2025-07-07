package models

type SendItem struct {
	Payload  string
	FileName string
	FilePath string
}

type ReceiveItem struct {
	Id       int
	Payload  string `json:"payload"`
	FileName string `json:"file_name"`
	Encoding string `json:"encoding"`
}
