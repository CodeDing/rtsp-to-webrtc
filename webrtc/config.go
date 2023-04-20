package webrtc

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Conf is a configuration.
type Conf struct {
	// general
	LogLevel                  LogLevel        `json:"logLevel"`
	LogDestinations           LogDestinations `json:"logDestinations"`
	LogFile                   string          `json:"logFile"`
	ReadTimeout               StringDuration  `json:"readTimeout"`
	WriteTimeout              StringDuration  `json:"writeTimeout"`
	ReadBufferCount           int             `json:"readBufferCount"`
	ExternalAuthenticationURL string          `json:"externalAuthenticationURL"`
	RemoteRtspAddress         string          `json:"remoteRtspAddress"`

	// WebRTC
	WebRTCDisable           bool       `json:"webrtcDisable"`
	WebRTCAddress           string     `json:"webrtcAddress"`
	WebRTCEncryption        bool       `json:"webrtcEncryption"`
	WebRTCServerKey         string     `json:"webrtcServerKey"`
	WebRTCServerCert        string     `json:"webrtcServerCert"`
	WebRTCAllowOrigin       string     `json:"webrtcAllowOrigin"`
	WebRTCTrustedProxies    IPsOrCIDRs `json:"webrtcTrustedProxies"`
	WebRTCICEServers        []string   `json:"webrtcICEServers"`
	WebRTCICEHostNAT1To1IPs []string   `json:"webrtcICEHostNAT1To1IPs"`
	WebRTCICEUDPMuxAddress  string     `json:"webrtcICEUDPMuxAddress"`
	WebRTCICETCPMuxAddress  string     `json:"webrtcICETCPMuxAddress"`
}

func LoadConfig(fpath string) (*Conf, error) {
	conf := &Conf{}
	err := loadFromFile(fpath, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func loadFromFile(fpath string, conf *Conf) error {
	byts, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}

	// load YAML config into a generic map
	var temp interface{}
	err = yaml.Unmarshal(byts, &temp)
	if err != nil {
		return err
	}

	// convert interface{} keys into string keys to avoid JSON errors
	var convert func(i interface{}) (interface{}, error)
	convert = func(i interface{}) (interface{}, error) {
		switch x := i.(type) {
		case map[interface{}]interface{}:
			m2 := map[string]interface{}{}
			for k, v := range x {
				ks, ok := k.(string)
				if !ok {
					return nil, fmt.Errorf("integer keys are not supported (%v)", k)
				}

				m2[ks], err = convert(v)
				if err != nil {
					return nil, err
				}
			}
			return m2, nil

		case []interface{}:
			a2 := make([]interface{}, len(x))
			for i, v := range x {
				a2[i], err = convert(v)
				if err != nil {
					return nil, err
				}
			}
			return a2, nil
		}

		return i, nil
	}
	temp, err = convert(temp)
	if err != nil {
		return err
	}

	// check for non-existent parameters
	var checkNonExistentFields func(what interface{}, ref interface{}) error
	checkNonExistentFields = func(what interface{}, ref interface{}) error {
		if what == nil {
			return nil
		}

		ma, ok := what.(map[string]interface{})
		if !ok {
			return fmt.Errorf("not a map")
		}

		for k, _ := range ma {
			fi := func() reflect.Type {
				rr := reflect.TypeOf(ref)
				for i := 0; i < rr.NumField(); i++ {
					f := rr.Field(i)
					if f.Tag.Get("json") == k {
						return f.Type
					}
				}
				return nil
			}()
			if fi == nil {
				return fmt.Errorf("non-existent parameter: '%s'", k)
			}

			// if fi == reflect.TypeOf(map[string]*PathConf{}) && v != nil {
			// 	ma2, ok := v.(map[string]interface{})
			// 	if !ok {
			// 		return fmt.Errorf("parameter %s is not a map", k)
			// 	}

			// 	for k2, v2 := range ma2 {
			// 		err := checkNonExistentFields(v2, reflect.Zero(fi.Elem().Elem()).Interface())
			// 		if err != nil {
			// 			return fmt.Errorf("parameter %s, key %s: %s", k, k2, err)
			// 		}
			// 	}
			// }
		}
		return nil
	}
	err = checkNonExistentFields(temp, Conf{})
	if err != nil {
		return err
	}

	// convert the generic map into JSON
	byts, err = json.Marshal(temp)
	if err != nil {
		return err
	}

	// load the configuration from JSON
	err = json.Unmarshal(byts, conf)
	if err != nil {
		return err
	}

	return nil
}

// IPsOrCIDRs is a parameter that contains a list of IPs or CIDRs.
type IPsOrCIDRs []fmt.Stringer

// MarshalJSON implements json.Marshaler.
func (d IPsOrCIDRs) MarshalJSON() ([]byte, error) {
	out := make([]string, len(d))

	for i, v := range d {
		out[i] = v.String()
	}

	sort.Strings(out)

	return json.Marshal(out)
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *IPsOrCIDRs) UnmarshalJSON(b []byte) error {
	var in []string
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}

	if len(in) == 0 {
		return nil
	}

	for _, t := range in {
		if _, ipnet, err := net.ParseCIDR(t); err == nil {
			*d = append(*d, ipnet)
		} else if ip := net.ParseIP(t); ip != nil {
			*d = append(*d, ip)
		} else {
			return fmt.Errorf("unable to parse IP/CIDR '%s'", t)
		}
	}

	return nil
}

// unmarshalEnv implements envUnmarshaler.
func (d *IPsOrCIDRs) unmarshalEnv(s string) error {
	byts, _ := json.Marshal(strings.Split(s, ","))
	return d.UnmarshalJSON(byts)
}
