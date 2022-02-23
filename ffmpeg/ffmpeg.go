package ffmpeg

import (
	"fmt"
	"os/exec"
)

func NewForwardStream(srcUrl, dstUrl string) error {
	cmdString := fmt.Sprintf("ffmpeg -i \"%s\" -c copy -f flv \"%s\"",
		srcUrl, dstUrl)
	cmd := exec.Command("sh", "-c", cmdString)

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	return nil
}
