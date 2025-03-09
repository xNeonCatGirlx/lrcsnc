package lrclib

import (
	"cmp"
	"encoding/json"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"lrcsnc/internal/log"
	"lrcsnc/internal/pkg/errors"
	"lrcsnc/internal/pkg/structs"
	"lrcsnc/internal/pkg/types"
)

var timeTagRegexp = regexp.MustCompile(`(\[[0-9]{2}:[0-9]{2}.[0-9]{2}])+`)

type DTO struct {
	Title        string  `json:"trackName"`
	Artist       string  `json:"artistName"`
	Album        string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

func dtoListToLyricsData(song structs.Song, bytes []byte) (structs.LyricsData, error) {
	dtos, err := parseDTOs(bytes)
	if err != nil {
		return structs.LyricsData{LyricsType: types.LyricsStateUnknown}, err
	}

	if len(dtos) == 0 {
		log.Debug("lrclib/dtoListToLyricsData", "Went in with empty DTO list. What's happening?")
		return structs.LyricsData{LyricsType: types.LyricsStateUnknown}, nil
	}

	dtos = removeMismatches(song, dtos)
	if len(dtos) > 1 {
		lyricsData := dtos[0].toLyricsData()
		return lyricsData, nil
	}

	return structs.LyricsData{LyricsType: types.LyricsStateNotFound}, errors.ErrLyricsNotFound
}

func parseDTOs(data []byte) ([]DTO, error) {
	var out DTO
	err := json.Unmarshal(data, &out)
	if err != nil {
		var outs []DTO
		err = json.Unmarshal(data, &outs)
		if err != nil {
			return nil, errors.ErrUnmarshalFail
		}
		return outs, nil
	}

	return []DTO{out}, nil
}

func (dto DTO) toLyricsData() (out structs.LyricsData) {
	if !dto.Instrumental && dto.PlainLyrics == "" && dto.SyncedLyrics == "" {
		out.LyricsType = types.LyricsStateUnknown
	}

	if dto.Instrumental {
		out.LyricsType = types.LyricsStateInstrumental
		return
	}

	if dto.PlainLyrics != "" && dto.SyncedLyrics == "" {
		lyrics := strings.Split(dto.PlainLyrics, "\n")
		for i := range out.Lyrics {
			lyrics[i] = sanitizeLyric(lyrics[i])
		}
		out.Lyrics = make([]structs.Lyric, 0, len(lyrics))
		for i, l := range lyrics {
			out.Lyrics[i] = structs.Lyric{Text: l}
		}

		out.LyricsType = types.LyricsStatePlain
		return
	}

	out.Lyrics = parseSyncedLyrics(dto.SyncedLyrics)
	out.LyricsType = types.LyricsStateSynced

	return
}

func parseSyncedLyrics(lyrics string) (out []structs.Lyric) {
	hasRepetitiveLyrics := false
	syncedLyrics := strings.Split(lyrics, "\n")

	out = make([]structs.Lyric, 0, len(syncedLyrics))

	for _, lyric := range syncedLyrics {
		timeTags := timeTagRegexp.FindAllString(lyric, -1)

		for _, ts := range timeTags {
			lyric = strings.Replace(lyric, ts, "", 1)
		}
		lyric = sanitizeLyric(lyric)

		hasRepetitiveLyrics = hasRepetitiveLyrics || len(timeTags) > 1

		for _, timeTagStr := range timeTags {
			timecode := parseTimeTag(timeTagStr)
			if timecode == -1 {
				continue
			}
			out = append(out, structs.Lyric{
				Time: timecode,
				Text: lyric,
			})
		}
	}

	if hasRepetitiveLyrics {
		slices.SortFunc(out, func(i, j structs.Lyric) int {
			return cmp.Compare(i.Time, j.Time)
		})
	}

	return
}

// A simple sanitize requires trimming any carriage return and space symbols
// It is wrapped into a function to be simple to update if needed
func sanitizeLyric(lyric string) string {
	return strings.TrimSpace(strings.TrimRight(lyric, "\r"))
}

// Returns the timestamp in seconds, specified in the provided timeTag
func parseTimeTag(timeTag string) float64 {
	// [01:23.45]
	if len(timeTag) != 10 {
		return -1
	}
	minutes, err := strconv.ParseFloat(timeTag[1:3], 64)
	if err != nil {
		return -1
	}
	seconds, err := strconv.ParseFloat(timeTag[4:9], 64)
	if err != nil {
		return -1
	}
	return minutes*60.0 + seconds
}

func removeMismatches(song structs.Song, lyricsData []DTO) []DTO {
	if len(lyricsData) == 0 {
		return lyricsData
	}

	var matchingLyricsData []DTO = make([]DTO, 0, len(lyricsData))

	for _, lyrics := range lyricsData {
		if strings.EqualFold(lyrics.Title, song.Title) &&
			// If player doesn't provide the song's duration, ignore the duration check
			// Otherwise, do a check that prevents different versions of a song of messing up the response
			((song.Duration != 0) == (math.Abs(float64(lyrics.Duration)-song.Duration) <= 2)) {
			matchingLyricsData = append(matchingLyricsData, lyrics)
		}
	}

	return matchingLyricsData
}
