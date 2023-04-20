# rtsp-to-webrtc
A stream protocol-translator for translating RTSP stream to WebRTC stream. The code is inherited from the open source [mediamtx](https://github.com/aler9/mediamtx), but it only take the advantage of providing the ability to read stream in WebRTC protocol natively.

## FFmpeg

```
$ffmpeg -re -stream_loop -1 -i test.mp4 -vf "settb=AVTB,
            setpts='trunc(PTS/1K)*1K+st(1,trunc(RTCTIME/1K))-1K*trunc(ld(1)/1K)',
            drawtext=fontsize=60:text='%{localtime}.%{eif\:1M*t-1K*trunc(t*1K)\:d}'" -x264-params "keyint=10:min-keyint=1" -crf 26 -c:v libx264 -an -b:v 500k -bufsize 500k -preset ultrafast  -rtsp_transport tcp -f rtsp rtsp://127.0.0.1:8554/mystream
```

```
# Config
  Stream server address is specified by `remoteRtspAddress` and the server's listen port is specified by `webrtcAddress: :9001`


# Build

```dotnetcli
$make build
```

# Run
```dotnetcli
./bin/rtsp-to-webrtc
```