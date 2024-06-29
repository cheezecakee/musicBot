package app

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var youtubeService *youtube.Service

func InitYouTube() error {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("error creating YouTube client: %v", err)
	}
	youtubeService = service
	return nil
}

func SearchVideo(ctx context.Context, query string) (string, error) {
	if youtubeService == nil {
		return "", fmt.Errorf("YoutTube client is not initiazlized")
	}
	call := youtubeService.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(1)
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
