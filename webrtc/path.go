package webrtc

import "github.com/aler9/gortsplib/pkg/base"

type Credential string

// TODO:
type path struct {
	name string

	ReadUser Credential
	ReadPass Credential
	ReadIPs  IPsOrCIDRs
}

type pathReaderSetupPlayRes struct {
	//path   *path
	stream *stream
	err    error
}

type pathDescribeRes struct {
	//path     *path
	stream *stream
	//redirect string
	err error
}

type pathErrAuthCritical struct {
	message  string
	response *base.Response
}

// Error implements the error interface.
func (pathErrAuthCritical) Error() string {
	return "critical authentication error"
}

type pathErrAuthNotCritical struct {
	message  string
	response *base.Response
}

// Error implements the error interface.
func (pathErrAuthNotCritical) Error() string {
	return "non-critical authentication error"
}
