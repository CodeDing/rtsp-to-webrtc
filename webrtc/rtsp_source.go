package webrtc

import (
	"context"
	"time"

	"github.com/CodeDing/rtsp-to-webrtc/formatprocessor"

	"github.com/aler9/gortsplib/v2"
	"github.com/aler9/gortsplib/v2/pkg/base"
	"github.com/aler9/gortsplib/v2/pkg/format"
	"github.com/aler9/gortsplib/v2/pkg/url"
	"github.com/pion/rtp"
)

func (c *webRTCConn) readRtspStream(ctx context.Context, streamReady chan<- struct{}) error {
	defer func() {
		close(streamReady)
	}()

	rc := gortsplib.Client{
		// Transport:       s.proto.Transport,
		// TLSConfig:       tlsConfig,
		// ReadTimeout:     time.Duration(s.readTimeout),
		// WriteTimeout:    time.Duration(s.writeTimeout),
		// ReadBufferCount: s.readBufferCount,
		// AnyPortEnable:   s.anyPortEnable,
		OnRequest: func(req *base.Request) {
			c.log(Debug, "c->s %v", req)
		},
		OnResponse: func(res *base.Response) {
			c.log(Debug, "s->c %v", res)
		},
		OnDecodeError: func(err error) {
			c.log(Warn, "%v", err)
		},
	}

	u, err := url.Parse(c.remoteRtspAddr + "/" + c.pathName)
	if err != nil {
		return err
	}

	err = rc.Start(u.Scheme, u.Host)
	if err != nil {
		return err
	}

	defer rc.Close()

	readErr := make(chan error)
	go func() {
		readErr <- func() error {
			medias, baseURL, _, err := rc.Describe(u)
			if err != nil {
				return err
			}

			err = rc.SetupAll(medias, baseURL)
			if err != nil {
				return err
			}

			// TODO: safe ???
			stream, err := newStream(medias, true)
			if err != nil {
				return err
			}

			c.stream = stream
			streamReady <- struct{}{}
			c.log(Info, "underlying rtsp stream is ready")

			defer func() {
				c.log(Info, "close rtsp stream")
				stream.close()
			}()

			for _, medi := range medias {
				for _, forma := range medi.Formats {
					cmedia := medi
					cformat := forma

					switch forma.(type) {
					case *format.H264:
						rc.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
							err := c.stream.writeData(cmedia, cformat, &formatprocessor.DataH264{
								RTPPackets: []*rtp.Packet{pkt},
								NTP:        time.Now(),
							})
							if err != nil {
								c.log(Warn, "receive RTP packet [%s] %v", forma, err)
							}
						})

					case *format.H265:
						rc.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
							err := c.stream.writeData(cmedia, cformat, &formatprocessor.DataH265{
								RTPPackets: []*rtp.Packet{pkt},
								NTP:        time.Now(),
							})
							if err != nil {
								c.log(Warn, "receive RTP packet [%s] %v", forma, err)
							}
						})

					case *format.VP8:
						rc.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
							err := c.stream.writeData(cmedia, cformat, &formatprocessor.DataVP8{
								RTPPackets: []*rtp.Packet{pkt},
								NTP:        time.Now(),
							})
							if err != nil {
								c.log(Warn, "receive RTP packet [%s] %v", forma, err)
							}
						})

					case *format.VP9:
						rc.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
							err := c.stream.writeData(cmedia, cformat, &formatprocessor.DataVP9{
								RTPPackets: []*rtp.Packet{pkt},
								NTP:        time.Now(),
							})
							if err != nil {
								c.log(Warn, "receive RTP packet [%s] %v", forma, err)
							}
						})

					case *format.MPEG4Audio:
						rc.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
							err := c.stream.writeData(cmedia, cformat, &formatprocessor.DataMPEG4Audio{
								RTPPackets: []*rtp.Packet{pkt},
								NTP:        time.Now(),
							})
							if err != nil {
								c.log(Warn, "receive RTP packet [%s] %v", forma, err)
							}
						})

					case *format.Opus:
						rc.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
							err := c.stream.writeData(cmedia, cformat, &formatprocessor.DataOpus{
								RTPPackets: []*rtp.Packet{pkt},
								NTP:        time.Now(),
							})
							if err != nil {
								c.log(Warn, "receive RTP packet [%s] %v", forma, err)
							}
						})

					default:
						rc.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
							err := c.stream.writeData(cmedia, cformat, &formatprocessor.DataGeneric{
								RTPPackets: []*rtp.Packet{pkt},
								NTP:        time.Now(),
							})
							if err != nil {
								c.log(Warn, "receive RTP packet [%s] %v", forma, err)
							}
						})
					}
				}
			}

			_, err = rc.Play(nil)
			if err != nil {
				return err
			}

			return rc.Wait()
		}()
	}()

	select {
	case err := <-readErr:
		return err

	case <-ctx.Done():
		rc.Close()
		<-readErr
		return nil
	}
}
