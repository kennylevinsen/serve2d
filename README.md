# serve2d
A protocol detecting server, based off the serve2 library.

You don't like having to have to decide what port to use for a service? Maybe you're annoyed by a firewall that only allows traffic to port 80? Or even a packet inspecting one that only allows real TLS traffic on port 443, but you want to SSH through none the less?

Welcome to serve2, a protocol recognizing and stacking server/dispatcher.

serve2 allows you to serve multiple protocols on a single socket. Example handlers include proxy, HTTP, TLS (through which HTTPS is handled), ECHO and DISCARD. More can easily be added, as long as the protocol sends some data that can be recognized. The proxy handler allows you to redirect the connection to external services, such as OpenSSH or Nginx, in case you don't want or can't use a Go implementation. In most cases, proxy will be sufficient.

So, what does this mean? Well, it means that if you run serve2 for port 22, 80 and 443 (or all the ports, although I would suggest just having your firewall redirect things in that case, rather than having 65535 listening sockets), you could ask for HTTP(S) on port 22, SSH on port 80, and SSH over TLS (Meaning undetectable without a MITM attack) on port 443! You have to admit that it's kind of neat.

All bytes read by serve2 are of course fed into whatever ends up having to handle the protocol. For more details on the protocol matching, look at the serve2 library directly.

# Installation
serve2d can be installed with:

      go get github.com/joushou/serve2d

It can be run with:

      cd $GOPATH/src/joushou/serve2d
      go build
      ./serve2d example_conf.json

Do note that the example_conf has the TLS protocol enabled, requiring a cert.pem and key.pem file (can be generated with http://golang.org/src/crypto/tls/generate_cert.go). To avoid this, simply remove the TLS entry.

# What's up with the name?
I called the first toy version "serve", and needed to call the new directory in my development folder something else, so it became serve2. 'd' was added to this configurable front-end (daemon), to distinguish it from the library.

# Usage
Due to potentially large amounts of parameters, serve2d consumes a json configuration file. The only parameter taken by serve2d is the name of this file. The format is as follows:

      {
         // Listening address as given directly to net.Listen.
         "address": ":80",

         // Maximum read size for protocol detection before fallback or failure.
         // Defaults to 128.
         "maxRead": 10,

         // Logging to stdout.
         // Defaults to false.
         "logStdout": true,

         // Logging to file. Note that only one logging destination can be
         // enabled at a given time.
         // Defaults to empty string, meaning disabled.
         "logFile": "serve2d.log",

         // The enabled ProtocolHandlers.
         "protocols": [
            {
               // Name of the ProtocolHandler.
               "kind": "proxy",

               // Setting this flag to true means that this ProtocolHandler
               // will not be used in protocol detection, but instead be used
               // as a fallback in case of failed detection.
               // Defaults to false.
               "default": false,

               // Protocol-specific configuration.
               // Defaults to empty.
               "conf": {
                  "magic": "SSH",
                  "target": "localhost:22"
               }
            }
         ]
      }

# ProtocolHandlers

## proxy
Simply dials another service to handle the protocol. Matches protocol using the user-defined string.

* magic (string): The bytes to look for in order to identify the protocol. Example: "SSH".
* target (string): The address as given directly to net.Dial to call the real service. Example: "localhost:22".

## tls
Looks for a TLS1.0-1.3 ClientHello handshake, and feeds it into Go's TLS handling. The resulting net.Conn is fed back through the protocol detection, allowing for any other supported protocol to go over TLS.
The certificates required can be generated with http://golang.org/src/crypto/tls/generate_cert.go.

* cert (string): The certificate PEM file path to use for the server. Example: "cert.pem".
* key (string): The key PEM file path to use for the server. Example: "key.pem".
* protos ([]string): The protocols the TLS server will advertise support for in the handshake. Example: ["http/1.1", "ssh"]

As tls works as a transport, it can be used for anything, not just HTTP. tls + proxy handler for SSH would make it possible to do the following to grant you stealthy SSH over TLS, which would be indistinguishable from HTTPS traffic (the named pipe/mkfifo is used to connect both stdin and stdout of netcat and openssl):

      mkfifo np
      nc -l 8888 < np | openssl s_client -connect hostwithserve2:443 -tls1 -quiet > np &
      ssh -p 8888 localhost

## http
Simple file-server without directory listing (might change in the future). It guards against navigating out of the directory with some simple path magic.

* path (string): Path to serve from. Example: "/srv/http/"
* notFoundMsg (string): 404 body. Example: "<!DOCTYPE html><html><body>Not Found</body></html>"

## echo
A test protocol. Requires that the client starts out by sending "ECHO" (which will by echoed by itself, of course). No configuration.

## discard
Same as DISCARD, start by sending "DISCARD". No configuration. If you feel silly, try DISCARD over TLS!

# More info
For more details about this project, see the underlying library: http://github.com/joushou/serve2
