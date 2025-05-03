package config

import (
	flag "github.com/spf13/pflag"
	"os"
	"strconv"
)

const (
	ReadTimeOut    int = 30
	WriteTimeOut   int = 30
	bootupWaitTime int = 5
)

type Configuration struct {
	Prefix         *string
	HttpPort       *string
	BootUpWaitTime *int
	WriteTimeOut   *int
	ReadTimeOut    *int
}

var (
	Config *Configuration

	prefix = flag.String(
		"prefix",
		"local",
		"Prefix")

	httpPort = flag.String(
		"http-port",
		"8080",
		"the port to serve on")

	bootupWaittime = flag.Int(
		"bootup-wait-time",
		bootupWaitTime,
		"Number of seconds to wait.")

	readTimeOut = flag.Int(
		"http-read-timeout",
		ReadTimeOut,
		"read time out")

	writeTimeOut = flag.Int(
		"http-write-timeout",
		WriteTimeOut,
		"write time out")
)

func updateStringEnvVariable(defValue *string, key string) *string {
	val := os.Getenv(key)

	if val == "" {
		return defValue
	}

	return &val
}

func updateIntEnvVariable(defValue *int, key string) *int {
	sVal := os.Getenv(key)
	if sVal == "" {
		return defValue
	}

	iVal, err := strconv.Atoi(sVal)
	if err != nil {
		return defValue
	}

	return &iVal
}

func init() {
	flag.Parse()

	prefix = updateStringEnvVariable(prefix, "BRAND_NAME")
	httpPort = updateStringEnvVariable(httpPort, "HTTP_PORT")
	writeTimeOut = updateIntEnvVariable(writeTimeOut, "WRITE_TIMEOUT")
	readTimeOut = updateIntEnvVariable(readTimeOut, "READ_TIMEOUT")

	Config = &Configuration{
		Prefix:         prefix,
		HttpPort:       httpPort,
		WriteTimeOut:   writeTimeOut,
		ReadTimeOut:    readTimeOut,
		BootUpWaitTime: bootupWaittime,
	}
}
