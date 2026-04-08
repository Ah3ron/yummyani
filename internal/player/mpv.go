// Package player provides an abstraction over external video players.
//
// The [Player] interface allows the TUI layer to launch playback
// without coupling to a specific player implementation.
package player

import (
	"fmt"
	"os"
	"os/exec"
)

// Player defines the interface for launching video playback.
type Player interface {
	Play(title, videoURL string) error
}

// MPV launches mpv with optimised flags for HTTP stream playback.
type MPV struct {
	Command string
}

// NewMPV creates an MPV player with the given binary name.
func NewMPV(command string) *MPV {
	if command == "" {
		command = "mpv"
	}
	return &MPV{Command: command}
}

// Play launches mpv and blocks until it exits.
func (m *MPV) Play(title, videoURL string) error {
	cmd := exec.Command(m.Command, buildMPVArgs(title, videoURL)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// buildMPVArgs returns mpv CLI flags optimised for HTTP stream playback.
func buildMPVArgs(title, videoURL string) []string {
	return []string{
		"--title=" + title,
		"--force-media-title=" + title,
		"--no-terminal",
		"--msg-level=all=error",
		"--force-seekable=yes",
		"--http-header-fields=Referer: https://kodikplayer.com/",
		"--vo=gpu",
		"--cache=yes",
		"--demuxer-max-bytes=150M",
		"--demuxer-max-back-bytes=75M",
		"--demuxer-readahead-secs=20",
		"--tls-verify=no",
		"--network-timeout=20",
		"--stream-lavf-o=analyzeduration=2000000",
		"--stream-lavf-o=probesize=5000000",
		"--demuxer-lavf-o=reconnect=1",
		"--demuxer-lavf-o=reconnect_streamed=1",
		"--video-sync=audio",
		"--framedrop=vo",
		"--vd-lavc-threads=0",
		"--deband=no",
		"--interpolation=no",
		"--sub-auto=no",
		"--osd-level=0",
		"--idle=no",
		videoURL,
	}
}

// LinkType normalises a MIME-type string to a short label.
func LinkType(typ string) string {
	switch typ {
	case "hls", "application/x-mpegURL":
		return "HLS"
	case "mp4", "video/mp4", "":
		return "MP4"
	default:
		if len(typ) > 0 {
			return typ
		}
		return "MP4"
	}
}

// FormatDuration converts seconds to "M:SS" format.
func FormatDuration(seconds int) string {
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}
