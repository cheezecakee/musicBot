package player

import (
	"fmt"
	"strings"
)

type Song struct {
	ID     int
	Name   string
	Artist string
	Url    string
}

type Queue struct {
	Songs   []Song
	Current int
}

// AddSong adds a song to the queue with an incremented ID
func (q *Queue) AddSong(song Song) {
	song.ID = len(q.Songs) + 1
	q.Songs = append(q.Songs, song)
}

func (q *Queue) Next() {
	if q.Current < len(q.Songs)-1 {
		q.Current++
	}
}

func (q *Queue) Previous() {
	if q.Current > 0 {
		q.Current--
	}
}

func (q *Queue) GetCurrentSong() *Song {
	if q.Current < len(q.Songs) {
		return &q.Songs[q.Current]
	}
	return nil
}

// Removes song by index number
func (q *Queue) RemoveSong(index int) {
	if index >= 0 && index < len(q.Songs) {
		q.Songs = append(q.Songs[:index], q.Songs[index+1:]...)
		if q.Current >= index {
			q.Current--
		}
	}
}

// Removes all songs from queue
func (q *Queue) Clear() {
	q.Songs = []Song{}
	q.Current = 0
}

func (q *Queue) String() string {
	if len(q.Songs) == 0 {
		return "The queue is empty."
	}

	var result strings.Builder
	for i, song := range q.Songs {
		if i == q.Current {
			result.WriteString(fmt.Sprintf("-> %d: %s by %s\n", i+1, song.Name, song.Artist))
		} else {
			result.WriteString(fmt.Sprintf("%d: %s by %s\n", i+1, song.Name, song.Artist))
		}
	}
	return result.String()
}
