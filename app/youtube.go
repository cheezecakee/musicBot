package app

import (
	"context"
	"fmt"
)

func SearchVideo(ctx context.Context, query string) (string, error) {
	call := Clients.Youtube.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(1)
	response, err := call.Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("error making Youtube API call: %v", err)
	}
	if len(response.Items) == 0 {
		return "", fmt.Errorf("no videos found")
	}
	videoID := response.Items[0].Id.VideoId
	if videoID == "" {
		return "", fmt.Errorf("video ID not found in API response")
	}
	return videoID, nil
}
