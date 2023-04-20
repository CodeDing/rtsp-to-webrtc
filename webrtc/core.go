package webrtc

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/alecthomas/kong"

	"github.com/gin-gonic/gin"
)

var version = "v0.0.0"

type Core struct {
	ctx          context.Context
	ctxCancel    func()
	confPath     string
	conf         *Conf
	logger       *Logger
	webRTCServer *webRTCServer

	done chan struct{}
}

var cli struct {
	Version  bool   `help:"print version"`
	Confpath string `arg:"" default:"rtsp-to-webrtc.yaml"`
}

func NewCore(args []string) (*Core, bool) {
	parser, err := kong.New(&cli,
		kong.Description("rtsp-to-webrtc "+version),
		kong.UsageOnError(),
		kong.ValueFormatter(func(value *kong.Value) string {
			switch value.Name {
			case "confpath":
				return "path to a config file. The default is rtsp-to-webrtc.yml."

			default:
				return kong.DefaultHelpValueFormatter(value)
			}
		}))
	if err != nil {
		panic(err)
	}

	_, err = parser.Parse(args)
	parser.FatalIfErrorf(err)

	if cli.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	gin.SetMode(gin.ReleaseMode)

	ctx, ctxCancel := context.WithCancel(context.Background())

	p := &Core{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		confPath:  cli.Confpath,
		done:      make(chan struct{}),
	}

	p.conf, err = LoadConfig(p.confPath)
	if err != nil {
		panic(err)
	}

	// TODO:
	p.logger, err = NewLogger(
		Level(p.conf.LogLevel),
		p.conf.LogDestinations,
		p.conf.LogFile,
	)
	if err != nil {
		panic(err)
	}

	p.Log(Debug, "Config: %v", *p.conf)

	if err := p.createResource(); err != nil {
		panic(err)
	}

	go p.run()

	return p, true
}

func (p *Core) createResource() error {
	// Hard code here, it may need to be improved
	var err error

	// Make the parameter configurable later
	p.webRTCServer, err = newWebRTCServer(
		p.ctx,
		p.conf.ExternalAuthenticationURL, // externalAuthURL
		p.conf.RemoteRtspAddress,         //
		p.conf.WebRTCAddress,             // Addresss
		p.conf.WebRTCEncryption,
		p.conf.WebRTCServerKey,
		p.conf.WebRTCServerCert,
		p.conf.WebRTCAllowOrigin,
		p.conf.WebRTCTrustedProxies,
		p.conf.WebRTCICEServers, //ice_servers [ip]
		p.conf.ReadBufferCount,  //read buf count
		p,
		p.conf.WebRTCICEHostNAT1To1IPs, // ICEHostNAT1To1IPs
		p.conf.WebRTCICEUDPMuxAddress,  // ICEUDPMuxAddress [:8189]
		p.conf.WebRTCICETCPMuxAddress,  // ICETCPMuxAddress
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *Core) run() {
	defer close(p.done)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

outer:
	for {
		select {
		case <-interrupt:
			p.Log(Info, "shutting down gracefully")
			break outer

		case <-p.ctx.Done():
			break outer
		}
	}

	p.ctxCancel()
	p.closeResource()
}

func (p *Core) closeResource() {
	p.webRTCServer.close()
	p.webRTCServer = nil

	p.logger.Close()
	p.logger = nil
}

// Log is the main logging function.
func (p *Core) Log(level Level, format string, args ...interface{}) {
	p.logger.Log(level, format, args...)
}

func (p *Core) Close() {
	p.ctxCancel()
	<-p.done
}

// Wait waits for the Core to exit.
func (p *Core) Wait() {
	<-p.done
}
