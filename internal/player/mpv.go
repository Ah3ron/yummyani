package player

import (
	"fmt"
	"os"
	"os/exec"
)

type Player interface {
	Play(title, videoURL string) error
}

type MPV struct {
	Command string
}

func NewMPV(command string) *MPV {
	if command == "" {
		command = "mpv"
	}
	return &MPV{Command: command}
}

func (m *MPV) Play(title, videoURL string) error {
	cmd := exec.Command(m.Command, buildMPVArgs(title, videoURL)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func buildMPVArgs(title, videoURL string) []string {
	return []string{
		"--title=" + title,
		"--force-media-title=" + title,
		"--terminal=no",
		"--force-seekable=yes",
		"--http-header-fields=Referer: https://kodikplayer.com/",
		"--cache=yes",
		"--demuxer-max-bytes=50M",
		"--demuxer-max-back-bytes=25M",
		"--demuxer-readahead-secs=10",
		"--network-timeout=10",
		videoURL,
	}
}

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

func FormatDuration(seconds int) string {
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}
