package main

import (
	"os"

	"github.com/CodeDing/rtsp-to-webrtc/webrtc"
)

func main() {
	s, ok := webrtc.NewCore(os.Args[1:])
	if !ok {
		os.Exit(1)
	}
	s.Wait()
}
