# net.tau - Network Library for Tau

A comprehensive networking library for Tau, inspired by Go's `net` package. Provides high-level networking primitives for TCP, UDP, and basic HTTP operations.

## Overview

The `net.tau` library provides a Go-style API for network programming in Tau:

- **Dial()** / **Listen()** - High-level connection and listener creation
- **Conn interface** - Represents network connections with Read/Write/Close methods
- **Listener interface** - Represents network listeners with Accept/Close methods
- **PacketConn interface** - Represents packet-oriented connections (UDP)

## Quick Start

### TCP Server

```tau
net = import("stdlib/net.tau")

# Listen for connections
ln = Listen("tcp", ":8080")

# Accept and handle connections
for {
    conn = ln.Accept()

    # Read data
    data = conn.Read(1024)

    # Write response
    conn.Write("Hello from Tau!")

    # Close connection
    conn.Close()
}

ln.Close()
```

### TCP Client

```tau
net = import("stdlib/net.tau")

# Connect to server
conn = Dial("tcp", "localhost:8080")

# Write data
conn.Write("Hello, server!")

# Read response
data = conn.Read(1024)
print(string(data))

# Close connection
conn.Close()
```

### UDP Server

```tau
net = import("stdlib/net.tau")

# Listen for packets
pc = ListenPacket("udp", ":9090")

# Receive packets
data = pc.ReadFrom(1024)
print(string(data))

pc.Close()
```

### UDP Client

```tau
net = import("stdlib/net.tau")

# Connect via UDP
conn = Dial("udp", "localhost:9090")

# Send data
conn.Write("UDP message")

conn.Close()
```

## API Reference

### Connection Functions

#### Dial(network, address)
Connect to a network address. Returns a Conn or error.

**Parameters:**
- `network` - "tcp", "tcp4", "udp", "udp4"
- `address` - "host:port" format

**Example:**
```tau
conn = Dial("tcp", "example.com:80")
```

**Go equivalent:** `conn, err := net.Dial("tcp", "example.com:80")`

#### DialTCP(network, address)
Connect to a TCP address. Convenience wrapper around Dial().

**Parameters:**
- `network` - "tcp" or "tcp4" (can be null, defaults to "tcp")
- `address` - "host:port" format

**Example:**
```tau
conn = DialTCP(null, "localhost:8080")
```

**Go equivalent:** `conn, err := net.DialTCP("tcp", nil, addr)`

#### DialUDP(network, address)
Connect to a UDP address. Convenience wrapper around Dial().

**Parameters:**
- `network` - "udp" or "udp4" (can be null, defaults to "udp")
- `address` - "host:port" format

**Example:**
```tau
conn = DialUDP(null, "localhost:9090")
```

**Go equivalent:** `conn, err := net.DialUDP("udp", nil, addr)`

### Listener Functions

#### Listen(network, address)
Create a network listener. Returns a Listener or error.

**Parameters:**
- `network` - "tcp", "tcp4", "tcp6"
- `address` - ":port" or "host:port" format

**Example:**
```tau
ln = Listen("tcp", ":8080")
```

**Go equivalent:** `ln, err := net.Listen("tcp", ":8080")`

#### ListenTCP(network, address)
Create a TCP listener. Convenience wrapper around Listen().

**Parameters:**
- `network` - "tcp" or "tcp4" (can be null, defaults to "tcp")
- `address` - ":port" or "host:port" format

**Example:**
```tau
ln = ListenTCP(null, ":8080")
```

**Go equivalent:** `ln, err := net.ListenTCP("tcp", addr)`

#### ListenPacket(network, address)
Create a packet-oriented listener. Returns a PacketConn or error.

**Parameters:**
- `network` - "udp", "udp4", "udp6"
- `address` - ":port" or "host:port" format

**Example:**
```tau
pc = ListenPacket("udp", ":9090")
```

**Go equivalent:** `pc, err := net.ListenPacket("udp", ":9090")`

#### ListenUDP(network, address)
Create a UDP listener. Convenience wrapper around ListenPacket().

**Parameters:**
- `network` - "udp" or "udp4" (can be null, defaults to "udp")
- `address` - ":port" or "host:port" format

**Example:**
```tau
pc = ListenUDP(null, ":9090")
```

**Go equivalent:** `conn, err := net.ListenUDP("udp", addr)`

### Conn Interface

A Conn represents a network connection and provides the following methods:

#### conn.Read(bufsize)
Read data from the connection. Returns bytes or error.

**Parameters:**
- `bufsize` - Maximum bytes to read (optional, defaults to 4096)

**Returns:** bytes or error. Empty bytes indicates EOF.

**Example:**
```tau
data = conn.Read(1024)
if !failed(data) {
    print(string(data))
}
```

**Go equivalent:** `n, err := conn.Read(buffer)`

#### conn.Write(data)
Write data to the connection. Returns number of bytes written or error.

**Parameters:**
- `data` - String or bytes to write

**Returns:** Number of bytes written, or error

**Example:**
```tau
n = conn.Write("Hello, World!")
```

**Go equivalent:** `n, err := conn.Write([]byte("Hello, World!"))`

#### conn.Close()
Close the connection. Returns null on success, error on failure.

**Example:**
```tau
err = conn.Close()
```

**Go equivalent:** `err := conn.Close()`

#### conn.LocalAddr()
Get the local network address.

**Returns:** Address string

**Go equivalent:** `addr := conn.LocalAddr()`

#### conn.RemoteAddr()
Get the remote network address.

**Returns:** Address string

**Go equivalent:** `addr := conn.RemoteAddr()`

### Listener Interface

A Listener represents a network listener and provides:

#### listener.Accept()
Accept the next connection. Returns Conn or error.

**Example:**
```tau
conn = ln.Accept()
if !failed(conn) {
    # Handle connection
}
```

**Go equivalent:** `conn, err := ln.Accept()`

#### listener.Close()
Close the listener. Returns null on success, error on failure.

**Example:**
```tau
ln.Close()
```

**Go equivalent:** `err := ln.Close()`

#### listener.Addr()
Get the listener's network address.

**Returns:** Address string

**Go equivalent:** `addr := ln.Addr()`

### PacketConn Interface

A PacketConn represents a packet-oriented connection and provides:

#### packetConn.ReadFrom(bufsize)
Read a packet. Returns bytes or error.

**Parameters:**
- `bufsize` - Maximum bytes to read (optional, defaults to 4096)

**Example:**
```tau
data = pc.ReadFrom(1024)
```

**Go equivalent:** `n, addr, err := pc.ReadFrom(buffer)`

#### packetConn.WriteTo(data, ip, port)
Write a packet to an address. Returns number of bytes written or error.

**Parameters:**
- `data` - String or bytes to write
- `ip` - Destination IP address
- `port` - Destination port

**Example:**
```tau
n = pc.WriteTo("Hello", "192.168.1.1", 9090)
```

**Go equivalent:** `n, err := pc.WriteTo([]byte("Hello"), addr)`

#### packetConn.Close()
Close the packet connection.

**Example:**
```tau
pc.Close()
```

**Go equivalent:** `err := pc.Close()`

#### packetConn.LocalAddr()
Get the local address.

**Returns:** Address string

**Go equivalent:** `addr := pc.LocalAddr()`

### Address Utilities

#### JoinHostPort(host, port)
Combine host and port into "host:port" format.

**Parameters:**
- `host` - Host string
- `port` - Port number or string

**Example:**
```tau
addr = JoinHostPort("localhost", 8080)  # "localhost:8080"
```

**Go equivalent:** `addr := net.JoinHostPort("localhost", "8080")`

#### SplitHostPort(hostport)
Split "host:port" into components.

**Parameters:**
- `hostport` - "host:port" string

**Returns:** Object with `Host` and `Port` fields, or error

**Example:**
```tau
result = SplitHostPort("localhost:8080")
print(result.Host)  # "localhost"
print(result.Port)  # "8080"
```

**Go equivalent:** `host, port, err := net.SplitHostPort("localhost:8080")`

#### ParseIP(s)
Parse an IP address string.

**Parameters:**
- `s` - IP address string

**Returns:** IP object or error

**Example:**
```tau
ip = ParseIP("192.168.1.1")
if !failed(ip) {
    print(ip.String())
}
```

**Go equivalent:** `ip := net.ParseIP("192.168.1.1")`

#### ResolveTCPAddr(network, address)
Resolve a TCP address.

**Parameters:**
- `network` - "tcp", "tcp4", etc.
- `address` - "host:port" string

**Returns:** Address object with IP, Port, and String() method

**Example:**
```tau
addr = ResolveTCPAddr("tcp", "localhost:8080")
print(addr.IP)
print(addr.Port)
```

**Go equivalent:** `addr, err := net.ResolveTCPAddr("tcp", "localhost:8080")`

#### ResolveUDPAddr(network, address)
Resolve a UDP address (same as ResolveTCPAddr).

**Go equivalent:** `addr, err := net.ResolveUDPAddr("udp", "localhost:9090")`

#### LookupHost(host)
Look up a hostname.

**Parameters:**
- `host` - Hostname to resolve

**Returns:** Result or error

**Example:**
```tau
addrs = LookupHost("example.com")
```

**Go equivalent:** `addrs, err := net.LookupHost("example.com")`

#### LookupPort(network, service)
Look up a port number for a service name.

**Parameters:**
- `network` - Network type (not currently used)
- `service` - Service name ("http", "https", "ssh", etc.)

**Returns:** Port number or error

**Example:**
```tau
port = LookupPort("tcp", "http")  # Returns 80
```

**Go equivalent:** `port, err := net.LookupPort("tcp", "http")`

### HTTP Helpers

#### HTTPGet(url)
Perform an HTTP GET request.

**Parameters:**
- `url` - HTTP URL (must start with "http://")

**Returns:** Response string or error

**Example:**
```tau
response = HTTPGet("http://example.com:80/")
print(response)
```

**Note:** This is a simplified helper. For full HTTP support, a dedicated http.tau library would be more appropriate (like Go's separate net/http package).

#### HTTPServe(conn, handler)
Handle an HTTP request on a connection.

**Parameters:**
- `conn` - Connection object from Accept()
- `handler` - Function that takes request object and returns response string

**Example:**
```tau
handler = fn(req) {
    body = "Hello!"
    return "HTTP/1.1 200 OK\r\nContent-Length: {len(body)}\r\n\r\n{body}"
}

ln = Listen("tcp", ":8080")
for {
    conn = ln.Accept()
    HTTPServe(conn, handler)
}
```

## Design Philosophy

This library follows Go's `net` package design philosophy:

1. **Simple, composable interfaces** - Conn, Listener, and PacketConn provide clean abstractions
2. **Consistent naming** - Functions match Go's naming conventions where possible
3. **Error handling** - Uses Tau's `failed()` builtin to check for errors
4. **High-level primitives** - Abstracts away low-level socket details
5. **Network type flexibility** - Supports "tcp", "tcp4", "udp", "udp4" like Go

## Comparison with Go

| Tau net.tau | Go net package |
|-------------|----------------|
| `Dial("tcp", "host:port")` | `net.Dial("tcp", "host:port")` |
| `Listen("tcp", ":8080")` | `net.Listen("tcp", ":8080")` |
| `conn.Read(bufsize)` | `conn.Read(buffer)` |
| `conn.Write(data)` | `conn.Write(data)` |
| `conn.Close()` | `conn.Close()` |
| `ln.Accept()` | `ln.Accept()` |
| `SplitHostPort(addr)` | `net.SplitHostPort(addr)` |
| `JoinHostPort(h, p)` | `net.JoinHostPort(h, p)` |

## Examples

See the following example files:
- `examples/net_go_style.tau` - Go-style networking patterns
- `examples/net_examples.tau` - Practical networking examples
- `examples/net_test.tau` - Test suite and examples

## Limitations

- IPv6 support is limited
- No support for Unix domain sockets yet
- No timeout or deadline support
- No TLS/SSL support
- HTTP helpers are basic (consider a separate http.tau library for full support)

## Future Enhancements

Potential additions to match more of Go's net package:

- Connection deadlines (SetDeadline, SetReadDeadline, SetWriteDeadline)
- TCP-specific options (SetKeepAlive, SetNoDelay)
- Unix domain socket support
- More complete DNS resolution
- Interface enumeration (net.Interfaces)
- IP network operations (net.IPNet)

## License

Part of the Tau programming language standard library.
