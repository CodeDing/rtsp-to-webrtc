package webrtc

import "os"

func writeTempFile(byts []byte) (string, error) {
	tmpf, err := os.CreateTemp(os.TempDir(), "rtsp-")
	if err != nil {
		return "", err
	}
	defer tmpf.Close()

	_, err = tmpf.Write(byts)
	if err != nil {
		return "", err
	}

	return tmpf.Name(), nil
}

func newInstance(conf string) (*Core, bool) {
	if conf == "" {
		return NewCore([]string{})
	}

	tmpf, err := writeTempFile([]byte(conf))
	if err != nil {
		return nil, false
	}
	defer os.Remove(tmpf)

	return NewCore([]string{tmpf})
}
