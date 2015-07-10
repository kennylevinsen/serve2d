package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/joushou/serve2"
	"github.com/joushou/serve2/proto"
)

var (
	server    *serve2.Server
	conf      Config
	confReady bool
	logger    func(format string, v ...interface{})
)

type Config struct {
	Address   string
	Logging   bool
	LogFile   string `json:"logFile,omitempty"`
	maxRead   int    `json:"maxRead,omitempty"`
	Protocols []Protocol
}

type Protocol struct {
	Kind      string
	AsDefault bool                   `json:"default,omitempty"`
	Conf      map[string]interface{} `json:"conf,omitempty"`
}

func logit(format string, msg ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			println("Log failed: ", r)
			panic(r)
		}
	}()

	if logger != nil || !confReady {
		log.Printf(format, msg...)
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			logit("Panicked: %s", err)
		}
	}()

	if len(os.Args) <= 1 {
		panic("Missing configuration path")
	}

	path := os.Args[1]

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		logit("Reading configuration failed")
		panic(err)
	}

	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		logit("Parsing configuration failed")
		panic(err)
	}

	confReady = true

	server = serve2.New()

	if conf.maxRead != 0 {
		server.BytesToCheck = conf.maxRead
	}

	if conf.Logging {
		if conf.LogFile != "" {
			file, err := os.Create(conf.LogFile)
			if err != nil {
				logit("Failed to open logfile: %s", conf.LogFile)
				panic(err)
			}
			log.SetOutput(file)
		}
		logger = log.Printf
		server.Logger = log.Printf
	}

	l, err := net.Listen("tcp", conf.Address)
	if err != nil {
		logit("Listen on [%s] failed", conf.Address)
		panic(err)
	}

	logit("Listening on: %s", conf.Address)

	for _, v := range conf.Protocols {
		var (
			handler serve2.ProtocolHandler
			err     error
		)
		switch v.Kind {
		case "proxy":
			magic, ok := v.Conf["magic"].(string)
			if !ok {
				panic("Proxy declaration is missing valid magic")
			}

			target, ok := v.Conf["target"].(string)
			if !ok {
				panic("Proxy declaration is missing valid target")
			}
			handler = proto.NewProxy([]byte(magic), "tcp", target)
		case "tls":
			cert, ok := v.Conf["cert"].(string)
			if !ok {
				panic("TLS decleration is missing valid certificate")
			}

			key, ok := v.Conf["key"].(string)
			if !ok {
				panic("TLS declaration is missing valid key")
			}

			protos, ok := v.Conf["protos"].([]string)
			if !ok {
				panic("TLS declaration is missing valid protocols")
			}

			handler, err = proto.NewTLS(protos, cert, key)
			if err != nil {
				logit("TLS configuration failed")
				panic(err)
			}
		case "echo":
			handler = proto.NewEcho()
		case "discard":
			handler = proto.NewDiscard()
		default:
			panic("Unknown kind: " + v.Kind)
		}

		if v.AsDefault {
			server.DefaultProtocol = handler
		} else {
			server.AddHandler(handler)
		}
	}

	server.Serve(l)
}
