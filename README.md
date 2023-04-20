# rtsp-to-webrtc
A stream protocol-translator for translating RTSP stream to WebRTC stream. The code is inherited from the open source [mediamtx](https://github.com/aler9/mediamtx), but it only take the advantage of providing the ability to read stream in WebRTC protocol natively.

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