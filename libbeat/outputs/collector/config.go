package collector

import (
	"time"

	"github.com/elastic/beats/v7/libbeat/common/transport/tlscommon"
)

type config struct {
	JobPath            string            `config:"job_path"`
	ContainerPath      string            `config:"container_path"`
	Params             map[string]string `config:"params"`
	Headers            map[string]string `config:"headers"`
	AuthUsername       string            `config:"auth_username"`
	AuthPassword       string            `config:"auth_password"`
	Method             string            `config:"method"`
	TLS                *tlscommon.Config `config:"ssl"`
	KeepAlive          time.Duration     `config:"keep_alive"`
	Timeout            time.Duration     `config:"timeout"`
	BulkMaxSize        int               `config:"bulk_max_size"`
	MaxRetries         int               `config:"max_retries"`
	Backoff            backoff           `config:"backoff"`
	LoadBalance        bool              `config:"load_balance"`
	CompressLevel      int               `config:"compress_level" validate:"min=0, max=9"`
	Limiter            limiterConfig     `config:"limiter"`
	BodyBytesPerSecond int               `config:"body_bytes_per_second"`
	BodyMaxBytes       int               `config:"body_max_bytes"`
	Output outputConfig `config:"output"`
}

type outputConfig struct {
	Params        map[string]string `config:"params"`
	Headers       map[string]string `config:"headers"`
	Method        string            `config:"method"`
	TLS           *tlscommon.Config `config:"ssl"`
	KeepAlive     time.Duration     `config:"keep_alive"`
	Timeout       time.Duration     `config:"timeout"`
	CompressLevel int               `config:"compress_level" validate:"min=0, max=9"`
}

type backoff struct {
	Init time.Duration `config:"init"`
	Max  time.Duration `config:"max"`
}

type limiterConfig struct {
	Quantity  int64         `config:"quantity"`
	Threshold int64         `config:"threshold"`
	Timeout   time.Duration `config:"timeout"`
}

var defaultConfig = config{
	JobPath:       "/collect/logs/job",
	ContainerPath: "/collect/logs/container",
	Method:        "POST",
	KeepAlive:     30 * time.Second,
	Timeout:       60 * time.Second,
	MaxRetries:    -1,
	Backoff: backoff{
		Init: 1 * time.Second,
		Max:  60 * time.Second,
	},
	LoadBalance:        true,
	CompressLevel:      9,
	BodyBytesPerSecond: 3145728,
	BodyMaxBytes:       5242880,
	Output: outputConfig{
		Method:        "POST",
		KeepAlive:     30 * time.Second,
		Timeout:       60 * time.Second,
		CompressLevel: 9,
	},
}

func (c *config) Validate() error {
	return nil
}
