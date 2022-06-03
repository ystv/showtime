package ffmpeg

import (
	"fmt"
	"os/exec"
)

// NewForwardStream takes a FFmpeg input and copies it to an RTMP URL.
func NewForwardStream(srcURL, dstURL string) error {
	cmdString := fmt.Sprintf("ffmpeg -i \"%s\" -c copy -f flv \"%s\"",
		srcURL, dstURL)
	cmd := exec.Command("sh", "-c", cmdString)

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	return nil
}

// NewVideoFromSingleImage creates a video file from a single image with a duration of 2 seconds.
func NewVideoFromSingleImage(srcPath, dstPath string) error {
	cmdString := fmt.Sprintf("ffmpeg -y -loop 1 -i \"%s\" -c:v libx264 -tune stillimage -t 2 -pix_fmt yuv420p -vf scale=1920:1080 \"%s\"",
		srcPath, dstPath)
	cmd := exec.Command("sh", "-c", cmdString)

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}
	return nil
}
