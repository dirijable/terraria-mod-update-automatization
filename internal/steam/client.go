package steam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchPublishedFileDetails(ctx context.Context, apiURL string, modIDs []string) (*SteamResponse, error) {
	formData := SetUrlValues(modIDs)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create steam request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("post mod ids to steam: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam api returned status: %s", resp.Status)
	}

	var steamResp SteamResponse
	if err := json.NewDecoder(resp.Body).Decode(&steamResp); err != nil {
		return nil, fmt.Errorf("parse steam json response: %w", err)
	}

	return &steamResp, nil
}