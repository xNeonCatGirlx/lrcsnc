package lyrics

import (
	"lrcsnc/internal/pkg/log"
	"lrcsnc/internal/pkg/structs"
	"lrcsnc/internal/romanization"
)

// Configure sets up the lyrics data by applying necessary configurations.
// Currently, it only applies romanization to the lyrics data.
// May be extended in the future.
//
// Every function/method/module/whatever needs to lock the mutex
// by themselves and only themselves.
// No locking a mutex in THIS function.
func Configure(lyricsData *structs.LyricsData) {
	log.Debug("lyrics/configure", "Starting configuring the received lyrics")

	// Romanization
	log.Debug("lyrics/configure", "Applying romanization if enabled and necessary")
	romanization.Romanize(lyricsData.Lyrics)

	log.Debug("lyrics/configure", "Done")
}
