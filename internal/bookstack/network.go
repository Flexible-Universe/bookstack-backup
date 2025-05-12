package bookstack

import (
	"fmt"
	"io"
	"net/http"
)

// httpGetWithAuth performs GET with BookStack token authentication using ID:Secret.
func (c *Client) httpGetWithAuth(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	authValue := fmt.Sprintf("Token %s:%s", c.inst.TokenID, c.inst.TokenSecret)
	req.Header.Set("Authorization", authValue)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
