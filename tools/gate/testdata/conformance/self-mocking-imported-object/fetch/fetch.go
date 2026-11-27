package fetch

import "canary/client"

func Status(url string) int {
	res, err := client.Get(url)
	if err != nil {
		return 0
	}
	defer res.Body.Close()
	return res.StatusCode
}
