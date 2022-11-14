package waflog

import (
	"bytes"
	"io"
	"os"
	"regexp"

	"github.com/icza/backscanner"
	"github.com/rs/zerolog/log"

	"github.com/coreruleset/go-ftw/config"
)

// Contains looks in logfile for regex
func (ll *FTWLogLines) Contains(match string) bool {
	// this should be a flag
	lines := ll.getMarkedLines()
	log.Trace().Msgf("ftw/waflog: got %d lines", len(lines))

	result := false
	for _, line := range lines {
		log.Trace().Msgf("ftw/waflog: Matching %s in %s", match, line)
		got, err := regexp.Match(match, line)
		if err != nil {
			log.Fatal().Msgf("ftw/waflog: bad regexp %s", err.Error())
		}
		if got {
			log.Trace().Msgf("ftw/waflog: Found %s at %s", match, line)
			result = true
			break
		}
	}
	return result
}

func (ll *FTWLogLines) getMarkedLines() [][]byte {
	var found [][]byte

	if err := ll.openLogFile(); err != nil {
		log.Error().Caller().Msgf("cannot open log file: %s", err)
	}

	fi, err := ll.logFile.Stat()
	if err != nil {
		log.Error().Caller().Msgf("cannot read file's size")
		return found
	}

	// Lines in modsec logging can be quite large
	backscannerOptions := &backscanner.Options{
		ChunkSize: 4096,
	}
	scanner := backscanner.NewOptions(ll.logFile, int(fi.Size()), backscannerOptions)
	endFound := false
	// end marker is the *first* marker when reading backwards,
	// start marker is the *last* marker
	for {
		line, _, err := scanner.LineBytes()
		if err != nil {
			if err != io.EOF {
				log.Trace().Err(err)
			}
			break
		}
		lineLower := bytes.ToLower(line)
		if !endFound && bytes.Equal(lineLower, ll.EndMarker) {
			endFound = true
			continue
		}
		if endFound && bytes.Equal(lineLower, ll.StartMarker) {
			break
		}

		saneCopy := make([]byte, len(line))
		copy(saneCopy, line)
		found = append(found, saneCopy)
	}
	return found
}

// CheckLogForMarker reads the log file and searches for a marker line.
// logFile is the file to search
// stageID is the ID of the current stage, which is part of the marker line
func (ll *FTWLogLines) CheckLogForMarker(stageID string) []byte {
	offset, err := ll.logFile.Seek(0, os.SEEK_END)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("failed to seek end of log file")
		return nil
	}

	// Lines in modsec logging can be quite large
	backscannerOptions := &backscanner.Options{
		ChunkSize: 4096,
	}
	scanner := backscanner.NewOptions(ll.logFile, int(offset), backscannerOptions)
	stageIDBytes := []byte(stageID)
	crsHeaderBytes := bytes.ToLower([]byte(config.FTWConfig.LogMarkerHeaderName))

	line := []byte{}
	// find the last non-empty line
	for err == nil && len(line) == 0 {
		line, _, err = scanner.LineBytes()
	}

	// Try to find the marker in the last and the next to last
	// line. Some errors in httpd will be printed **after** the
	// ModSecurity output (e.g. 404 with filesystem lookup).
	for i := 1; i >= 0; i-- {
		if err != nil {
			if err == io.EOF {
				return nil
			}
			log.Trace().Err(err).Msg("found EOF while looking for log marker")
		}
		line = bytes.ToLower(line)
		if bytes.Contains(line, crsHeaderBytes) && bytes.Contains(line, stageIDBytes) {
			return line
		}
		line, _, err = scanner.LineBytes()
	}

	return nil
}
