package main

import (
	"fmt"
	"net/url"
	"strconv"
)

type FileDetails struct {
	PublishedFileId string `json:"publishedfileid"`
	TimeUpdated     int64  `json:"time_updated"`
	Result          int    `json:"result"`
}

type Response struct {
	FileDetails []FileDetails `json:"publishedfiledetails"`
}

type SteamResponse struct {
	Response `json:"response"`
}

func SetUrlValues(modIds []string) url.Values {
	formData := url.Values{}
	formData.Set("itemcount", strconv.Itoa(len(modIds)))
	for idx, modId := range modIds {
		key := fmt.Sprintf("publishedfileids[%d]", idx)
		formData.Set(key, modId)
	}
	return formData
}
