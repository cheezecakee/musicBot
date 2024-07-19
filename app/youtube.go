package app

import (
	"fmt"
)

func SearchVideo(query string) (string, error) {
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

// GetTrackInfoFromYouTubeURL retrieves the title of a YouTube video based on its ID
func GetTrackInfoFromYouTubeURL(videoID string) (string, error) {
	// Fetch video details using the YouTube Data API
	call := Clients.Youtube.Videos.List([]string{"snippet"}).Id(videoID)
	response, err := call.Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("error fetching video details: %v", err)
	}

	if len(response.Items) == 0 {
		return "", fmt.Errorf("no video found for the given ID")
	}

	item := response.Items[0]
	title := item.Snippet.Title

	return title, nil
}
