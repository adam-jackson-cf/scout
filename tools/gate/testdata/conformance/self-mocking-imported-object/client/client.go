package client

import "net/http"

var Default = &http.Client{}

func Get(url string) (*http.Response, error) { return Default.Get(url) }
