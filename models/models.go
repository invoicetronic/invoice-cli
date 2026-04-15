package models

type SendItem struct {
	Payload  string `json:"payload"`
	FileName string `json:"file_name"`
	FilePath string `json:"-"`
}

type ReceiveItem struct {
	Id       int
	Payload  string `json:"payload"`
	FileName string `json:"file_name"`
	Encoding string `json:"encoding"`
}
