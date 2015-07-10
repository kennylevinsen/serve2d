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

   go build github.com/joushou/serve2d
   go install github.com/joushou/serve2d
   serve2d conf.json

# What's up with the name?
I called the first toy version "serve", and needed to call the new directory in my development folder something else, so it became serve2. 'd' was added to this configurable front-end (daemon), to distinguish it from the library.

# Usage
Due to potentially large amounts of parameters, serve2d consumes a json configuration file. The only parameter taken by serve2d is the name of this file. The format is as follows:

   {
      // Listening address as given directly to net.Listen
      "address": ":80",

      // Whether or not to log things (optional, defaults to false)
      "logging": true,

      // If logging to file is wanted rather than stdout, set this to the
      // wanted filename (optional, defaults to logging to stdout)
      "logFile": "serve2d.log", // Filename if log t

      // The enabled ProtocolHandlers
      "protocols": [
         {
            // Name of the ProtocolHandler
            "kind": "proxy",

            // If this ProtocolHandler should not detect the protocol, but
            // rather just be a fallback
            "default": false,

            // Protocol-specific configuration
            "conf": {
               "magic": "SSH",
               "target": "localhost:22"
            }
         }
      ]
   }

# Available ProtocolHandlers

## proxy
Simply dials another service to handle the protocol. Matches protocol using the user-defined string.

### Configuration
magic (string): The bytes to look for in order to identify the protocol
target (string): The address as given directly to net.Dial to call the real service.

## tls
Looks for a TLS1.0-1.3 ClientHello handshake, and feeds it into Go's TLS handling. The resulting net.Conn is fed back through the protocol detection, allowing for any other supported protocol to go over TLS.

### Configuration
cert (string): The certificate PEM file path to use for the server
key (string): The key PEM file path to use for the server
protos ([]string): The protocols the TLS server will advertise support for in the handshake.

## echo
A test protocol. Requires that the client starts out by sending "ECHO" (which will by echoed by itself, of course). No configuration.

## discard
Same as DISCARD, start by sending "DISCARD". Not configuration. If you feel silly, try DISCARD over TLS!

# More info
For more details about this project, see the underlying library:

   github.com/joushou/serve2
