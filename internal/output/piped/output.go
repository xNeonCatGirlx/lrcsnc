package piped

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"lrcsnc/internal/pkg/global"
	"lrcsnc/internal/pkg/log"
	"lrcsnc/internal/pkg/types"

	mprislib "github.com/Endg4meZer0/go-mpris"
)

var outputPath = "/dev/stdout"
var outputDestination = os.Stdout

// tempFile just adds .{pid}.tmp to the outputPath
var tempFile string

func outputIsStd() bool {
	return outputDestination == os.Stdout ||
		outputDestination == os.Stderr ||
		outputDestination == os.Stdin
}

var writeChan = make(chan string, 1)
var overwrite = ""
var pendingLyricIndex = -1
var instrumentalTimer *time.Timer = time.NewTimer(5 * time.Minute)

// Init initializes... basically everything.
func Init() {
	// Check for config's output destination
	global.Config.M.Lock()
	if global.Config.C.Output.Piped.Destination != "stdout" {
		outputPath = global.Config.C.Output.Piped.Destination
		changeOutput(outputPath)
	}
	global.Config.M.Unlock()

	// Initialize the writer chain
	go func() {
		for {
			s := <-writeChan
			Write(s)
		}
	}()

	// Initialize lyric change listener
	go func() {
		for {
			lyricIndex := <-currentLyricChangedChan
			if overwrite != "" {
				pendingLyricIndex = lyricIndex
				continue
			}
			lyric := FormatLyric(lyricIndex)
			if lyric == "" {
				instrumentalTimer.Reset(1)
			} else {
				instrumentalTimer.Stop()
				writeChan <- lyric
			}
		}
	}()

	// Initialize instrumental timer listener
	go func() {
		i := 1
		for {
			<-instrumentalTimer.C
			global.Config.M.Lock()
			global.Player.M.Lock()

			note := global.Config.C.Output.Piped.Instrumental.Symbol
			j := int(global.Config.C.Output.Piped.Instrumental.MaxSymbols + 1)

			// Only update instrumental stuff if the song is playing
			if global.Player.P.PlaybackStatus == mprislib.PlaybackPlaying {
				var stringToPrint string

				switch global.Player.P.Song.LyricsData.LyricsState {
				case types.LyricsStateSynced, types.LyricsStateInstrumental:
					stringToPrint = getInstrumentalString(global.Config.C.Output.Piped.Lyric, global.Config.C.Output.Piped.Text.Format)
				case types.LyricsStatePlain:
					stringToPrint = getInstrumentalMessage(global.Config.C.Output.Piped.NoSyncedLyrics, global.Config.C.Output.Piped.Text.Format)
				case types.LyricsStateNotFound:
					stringToPrint = getInstrumentalMessage(global.Config.C.Output.Piped.SongNotFound, global.Config.C.Output.Piped.Text.Format)
				case types.LyricsStateInProgress:
					stringToPrint = getInstrumentalMessage(global.Config.C.Output.Piped.GettingLyrics, global.Config.C.Output.Piped.Text.Format)
				default:
					stringToPrint = getInstrumentalMessage(global.Config.C.Output.Piped.ErrorMessage, global.Config.C.Output.Piped.Text.Format)
				}

				if len(stringToPrint) != 0 {
					stringToPrint += " "
				}
				stringToPrint += strings.Repeat(note, i%j)

				writeChan <- stringToPrint

				i++
				if i >= j {
					i = 1
				}
			} else {
				writeChan <- strings.TrimSpace(global.Config.C.Output.Piped.NotPlaying.Text)
				global.Player.M.Unlock()
				global.Config.M.Unlock()
				instrumentalTimer.Stop()
				continue
			}
			global.Player.M.Unlock()
			instrumentalTimer.Reset(time.Duration(global.Config.C.Output.Piped.Instrumental.Interval*1000) * time.Millisecond)
			global.Config.M.Unlock()
		}
	}()
}

// Write writes the string s to outputDestination.
// If the outputDestination is not related to std,
// then does its best to ensure the write is an atomic operation by using temp files.
// If JSON output is used, it will be formatted as JSON with full data.
func Write(s string) {
	global.Config.M.Lock()
	if global.Config.C.Output.Piped.JSON {
		s = FormatToJSON(s)
	}
	if global.Config.C.Output.Piped.InsertNewline {
		s = s + "\n"
	}
	global.Config.M.Unlock()

	if !outputIsStd() {
		if tempDestination, err := os.OpenFile(tempFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); err == nil {
			// Atomic copy (for better support of something like obs-text-pthread)
			tempDestination.Truncate(0)
			tempDestination.Seek(0, 0)
			tempDestination.WriteString(s)
			err = os.Rename(tempFile, outputPath)
			if err != nil {
				log.Error("output/piped", "Failed to move the temp file onto output destination: "+err.Error())
				return
			}
		} else {
			// If temp destination is unavailable, revert to basic handling (not atomic)
			outputDestination.Truncate(0)
			outputDestination.Seek(0, 0)
			outputDestination.WriteString(s)
		}
	} else {
		outputDestination.WriteString(s)
	}
}

// FormatLyric formats the lyric string (that is found by lyricIndex) to be displayed
// in accordance with the text format configuration.
func FormatLyric(lyricIndex int) string {
	global.Config.M.Lock()
	defer global.Config.M.Unlock()
	global.Player.M.Lock()
	defer global.Player.M.Unlock()

	lyric := lyricIndexToString(lyricIndex, global.Player.P.Song.LyricsData.Lyrics)
	if strings.TrimSpace(lyric) == "" {
		return ""
	}

	multiplierValue := 0
	for i := lyricIndex; i >= 0 && global.Player.P.Song.LyricsData.Lyrics[i].Text == lyric; i-- {
		multiplierValue++
	}
	multiplier := ""
	if multiplierValue > 1 {
		multiplier = strings.ReplaceAll(global.Config.C.Output.Piped.Multiplier.Format, "{value}", strconv.Itoa(multiplierValue))
	}
	replacer := strings.NewReplacer(
		"{icon}", global.Config.C.Output.Piped.Lyric.Icon,
		"{lyric}", lyric,
		"{multiplier}", multiplier,
	)
	return strings.TrimSpace(replacer.Replace(global.Config.C.Output.Piped.Text.Format))
}

// Overwrite sets the overwrite string to be displayed.
// Clears itself in 5 seconds.
func Overwrite(s string) {
	overwrite = s
	writeChan <- overwrite
	go func() {
		<-time.NewTimer(5 * time.Second).C
		overwriteEnds()
	}()
}

func overwriteEnds() {
	overwrite = ""
	currentLyricChangedChan <- pendingLyricIndex
	pendingLyricIndex = -1
}

// changeOutput changes the output destination to the specified path.
// The write check is usually performed at config validation step,
// but it's good to have it here too.
func changeOutput(p string) error {
	newDest, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Error("output/piped", "Error opening new output destination file, ignoring. More: "+err.Error())
		return err
	} else {
		r := strings.NewReplacer(
			"{pid}", strconv.Itoa(os.Getpid()),
		)
		tempFile = r.Replace(outputPath + ".{pid}.tmp")
		// We'll try to open temp file here to see if it even works
		_, err := os.OpenFile(tempFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Error("output/piped", fmt.Sprintf("Failed to open a temp write file (%s). The writes will not be atomic.", tempFile))
		}

		outputDestination = newDest
	}

	return nil
}

// Close closes the output if it is not related to std.
func Close() {
	if outputDestination != nil && !outputIsStd() {
		outputDestination.Close()
	}
}
