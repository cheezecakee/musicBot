package player

import (
	"fmt"
	"strings"
)

type Song struct {
	Name   string
	Artist string
	URL    string
}

type Queue struct {
	Songs   []Song
	Current int
}

func (q *Queue) AddSong(song Song) {
	q.Songs = append(q.Songs, song)
}

func (q *Queue) Next() {
	if q.Current < len(q.Songs)-1 {
		q.Current++
	}
}

func (q *Queue) Previous() {
	if q.Current < 0 {
		q.Current--
	}
}

func (q *Queue) CurrentSong() *Song {
	if len(q.Songs) == 0 {
		return nil
	}
	return &q.Songs[q.Current]
}

func (q *Queue) RemoveSong(index int) {
	if index < 0 || index >= len(q.Songs) {
		return
	}
	q.Songs = append(q.Songs[:index], q.Songs[index+1:]...)
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
