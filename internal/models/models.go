package models

// SABQueue represents the SabNZBD queue response
type SABQueue struct {
	Queue struct {
		MBLeft string `json:"mbleft"`
	} `json:"queue"`
}
