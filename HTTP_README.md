# http.tau - HTTP Client and Server Library for Tau

A comprehensive HTTP library for Tau, inspired by Go's `net/http` package. Provides both client and server functionality with a clean, idiomatic API.

## Overview

The `http.tau` library provides Go-style HTTP primitives:

- **HTTP Client** - Get(), Post(), NewClient(), Client.Do()
- **HTTP Server** - ListenAndServe(), NewServer()
- **Request Routing** - ServeMux, Handle(), HandleFunc()
- **Response Writing** - ResponseWriter interface
- **Status Codes** - StatusOK, StatusNotFound, etc.

## Quick Start

### Simple HTTP Server

```tau
http = import("net/http")

# Register a handler
http.HandleFunc("/", fn(w, req) {
    w.Write("Hello, World!")
})

# Start server
http.ListenAndServe(":8080", null)
```

**Go equivalent:**
```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
})
http.ListenAndServe(":8080", nil)
```

### HTTP Client

```tau
http = import("net/http")

# Make a GET request
resp = Get("http://example.com:80/")
if !failed(resp) {
    print(resp.Body)
}
```

**Go equivalent:**
```go
resp, _ := http.Get("http://example.com/")
// handle response
```

## API Reference

### HTTP Client

#### Get(url)
Issues a GET request to the specified URL.

**Parameters:**
- `url` - URL string (must start with "http://")

**Returns:** Response object or error

**Example:**
```tau
resp = Get("http://example.com:80/")
if !failed(resp) {
    print("Status: {resp.StatusCode}")
    print("Body: {resp.Body}")
}
```

**Go equivalent:** `resp, err := http.Get(url)`

#### Post(url, contentType, body)
Issues a POST request to the specified URL.

**Parameters:**
- `url` - URL string
- `contentType` - Content-Type header value
- `body` - Request body as string

**Returns:** Response object or error

**Example:**
```tau
resp = Post("http://api.example.com:80/users", "application/json", '{"name":"John"}')
```

**Go equivalent:** `resp, err := http.Post(url, contentType, body)`

#### NewClient()
Creates a new HTTP client.

**Returns:** Client object

**Example:**
```tau
client = NewClient()
resp = client.Get("http://example.com:80/")
```

**Go equivalent:** `client := &http.Client{}`

#### Client.Do(req)
Sends an HTTP request and returns a response.

**Parameters:**
- `req` - Request object created with NewRequest()

**Example:**
```tau
client = NewClient()
req = NewRequest("GET", "http://example.com:80/", null)
req.Header["user-agent"] = "TauHTTP/1.0"
resp = client.Do(req)
```

**Go equivalent:** `resp, err := client.Do(req)`

#### NewRequest(method, url, body)
Creates a new HTTP request.

**Parameters:**
- `method` - HTTP method ("GET", "POST", etc.)
- `url` - URL string
- `body` - Request body (can be null)

**Returns:** Request object

**Example:**
```tau
req = NewRequest("POST", "http://api.example.com:80/data", '{"key":"value"}')
req.Header["content-type"] = "application/json"
```

**Go equivalent:** `req, _ := http.NewRequest(method, url, body)`

### HTTP Server

#### ListenAndServe(addr, handler)
Listens on the TCP network address and serves HTTP requests.

**Parameters:**
- `addr` - Address to listen on (e.g., ":8080")
- `handler` - Handler to invoke (use `null` for default ServeMux)

**Example:**
```tau
ListenAndServe(":8080", null)
```

**Go equivalent:** `http.ListenAndServe(":8080", nil)`

#### HandleFunc(pattern, handler)
Registers a handler function for the given pattern on the default ServeMux.

**Parameters:**
- `pattern` - URL pattern (e.g., "/", "/api/")
- `handler` - Function(w, req) to handle requests

**Example:**
```tau
HandleFunc("/hello", fn(w, req) {
    w.Write("Hello!")
})
```

**Go equivalent:** `http.HandleFunc("/hello", handler)`

#### Handle(pattern, handler)
Registers a handler for the given pattern on the default ServeMux.

**Parameters:**
- `pattern` - URL pattern
- `handler` - Handler object with ServeHTTP(w, req) method

**Go equivalent:** `http.Handle(pattern, handler)`

#### NewServer(addr, handler)
Creates a new HTTP server.

**Parameters:**
- `addr` - Address to listen on
- `handler` - Request handler

**Returns:** Server object

**Example:**
```tau
mux = NewServeMux()
mux.HandleFunc("/", fn(w, req) { w.Write("Home") })

server = NewServer(":8080", mux)
server.ListenAndServe()
```

**Go equivalent:**
```go
server := &http.Server{Addr: ":8080", Handler: mux}
server.ListenAndServe()
```

### ServeMux (Request Multiplexer)

#### NewServeMux()
Creates a new request multiplexer for routing.

**Returns:** ServeMux object

**Example:**
```tau
mux = NewServeMux()
mux.HandleFunc("/", homeHandler)
mux.HandleFunc("/api/", apiHandler)
ListenAndServe(":8080", mux)
```

**Go equivalent:** `mux := http.NewServeMux()`

#### ServeMux.HandleFunc(pattern, handler)
Registers a handler function for the given pattern.

**Parameters:**
- `pattern` - URL pattern to match
- `handler` - Function(w, req) to handle matching requests

**Example:**
```tau
mux = NewServeMux()
mux.HandleFunc("/users", fn(w, req) {
    w.Header()["content-type"] = "application/json"
    w.Write('[{"id":1,"name":"Alice"}]')
})
```

**Go equivalent:** `mux.HandleFunc(pattern, handler)`

#### ServeMux.Handle(pattern, handler)
Registers a handler for the given pattern.

**Go equivalent:** `mux.Handle(pattern, handler)`

### ResponseWriter Interface

The ResponseWriter interface is used to construct HTTP responses.

#### w.Header()
Returns the header map.

**Example:**
```tau
w.Header()["content-type"] = "application/json"
w.Header()["x-custom-header"] = "value"
```

**Go equivalent:** `w.Header().Set("Content-Type", "application/json")`

#### w.Write(data)
Writes data to the response body.

**Parameters:**
- `data` - String or bytes to write

**Returns:** Number of bytes written

**Example:**
```tau
w.Write("Hello, ")
w.Write("World!")
```

**Go equivalent:** `w.Write([]byte("Hello, World!"))`

#### w.WriteHeader(statusCode)
Sends an HTTP response header with the provided status code.

**Parameters:**
- `statusCode` - HTTP status code (e.g., StatusOK, StatusNotFound)

**Example:**
```tau
w.WriteHeader(StatusNotFound)
w.Write("404 page not found")
```

**Go equivalent:** `w.WriteHeader(http.StatusNotFound)`

### Request Object

Request objects contain information about HTTP requests.

**Fields:**
- `Method` - HTTP method (e.g., "GET", "POST")
- `URL` - Request URL path
- `Proto` - Protocol version (e.g., "HTTP/1.1")
- `Header` - Request headers (map)
- `Body` - Request body (string)

**Example:**
```tau
HandleFunc("/api", fn(w, req) {
    if req.Method == MethodPost {
        data = req.Body
        # Process POST data
    }
})
```

### Response Object

Response objects contain HTTP response information.

**Fields:**
- `StatusCode` - HTTP status code (int)
- `Status` - Status line (string)
- `Proto` - Protocol version
- `Header` - Response headers (map)
- `Body` - Response body (string)

**Example:**
```tau
resp = Get("http://example.com:80/")
if resp.StatusCode == StatusOK {
    print(resp.Body)
}
```

### Status Codes

HTTP status code constants (matching Go's http package):

**2xx Success:**
- `StatusOK` - 200
- `StatusCreated` - 201
- `StatusAccepted` - 202
- `StatusNoContent` - 204

**3xx Redirection:**
- `StatusMovedPermanently` - 301
- `StatusFound` - 302
- `StatusSeeOther` - 303
- `StatusNotModified` - 304
- `StatusTemporaryRedirect` - 307
- `StatusPermanentRedirect` - 308

**4xx Client Errors:**
- `StatusBadRequest` - 400
- `StatusUnauthorized` - 401
- `StatusForbidden` - 403
- `StatusNotFound` - 404
- `StatusMethodNotAllowed` - 405
- `StatusRequestTimeout` - 408
- `StatusConflict` - 409
- `StatusGone` - 410

**5xx Server Errors:**
- `StatusInternalServerError` - 500
- `StatusNotImplemented` - 501
- `StatusBadGateway` - 502
- `StatusServiceUnavailable` - 503
- `StatusGatewayTimeout` - 504

### HTTP Methods

HTTP method constants:
- `MethodGet` - "GET"
- `MethodPost` - "POST"
- `MethodPut` - "PUT"
- `MethodDelete` - "DELETE"
- `MethodPatch` - "PATCH"
- `MethodHead` - "HEAD"
- `MethodOptions` - "OPTIONS"

### Helper Functions

#### StatusText(code)
Returns the text description for an HTTP status code.

**Example:**
```tau
text = StatusText(200)  # "OK"
text = StatusText(404)  # "Not Found"
```

**Go equivalent:** `http.StatusText(code)`

#### Error(w, error, code)
Replies to the request with the specified error message and HTTP code.

**Example:**
```tau
Error(w, "Bad Request", StatusBadRequest)
```

**Go equivalent:** `http.Error(w, error, code)`

#### Redirect(w, req, url, code)
Replies with a redirect to the specified URL.

**Example:**
```tau
Redirect(w, req, "/new-location", StatusFound)
```

**Go equivalent:** `http.Redirect(w, r, url, code)`

#### NotFoundHandler()
Returns a simple 404 handler.

**Example:**
```tau
mux.Handle("/old-path", NotFoundHandler())
```

**Go equivalent:** `http.NotFoundHandler()`

#### RedirectHandler(url, code)
Returns a handler that redirects to the given URL.

**Example:**
```tau
mux.Handle("/old", RedirectHandler("/new", StatusMovedPermanently))
```

**Go equivalent:** `http.RedirectHandler(url, code)`

#### FileServer(root)
Serves static files from a directory.

**Parameters:**
- `root` - Root directory path

**Example:**
```tau
mux.Handle("/static/", FileServer("./public"))
```

**Go equivalent:** `http.FileServer(http.Dir("./public"))`

## Complete Examples

### REST API Server

```tau
http = import("net/http")

mux = NewServeMux()

# GET /api/users
mux.HandleFunc("/api/users", fn(w, req) {
    if req.Method == MethodGet {
        w.Header()["content-type"] = "application/json"
        w.Write('[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]')
    } else if req.Method == MethodPost {
        w.WriteHeader(StatusCreated)
        w.Header()["content-type"] = "application/json"
        w.Write('{"id":3,"name":"Charlie"}')
    } else {
        w.WriteHeader(StatusMethodNotAllowed)
    }
})

# GET /health
mux.HandleFunc("/health", fn(w, req) {
    w.Header()["content-type"] = "application/json"
    w.Write('{"status":"healthy"}')
})

ListenAndServe(":8080", mux)
```

### HTTP Client with Custom Headers

```tau
http = import("net/http")

client = NewClient()

req = NewRequest("GET", "http://api.example.com:80/data", null)
req.Header["authorization"] = "Bearer token123"
req.Header["user-agent"] = "TauHTTP/1.0"

resp = client.Do(req)

if !failed(resp) {
    print("Status: {resp.StatusCode}")
    print("Body: {resp.Body}")
}
```

### Static File Server

```tau
http = import("net/http")

mux = NewServeMux()

# Serve static files from ./public
mux.Handle("/static/", FileServer("./public"))

# Home page
mux.HandleFunc("/", fn(w, req) {
    w.Header()["content-type"] = "text/html"
    w.Write("<h1>Welcome</h1><p>See <a href='/static/'>files</a></p>")
})

ListenAndServe(":8080", mux)
```

### Middleware Pattern

```tau
http = import("net/http")

# Logging middleware
loggingMiddleware = fn(next) {
    return fn(w, req) {
        print("[{req.Method}] {req.URL}")
        next(w, req)
    }
}

mux = NewServeMux()

handler = fn(w, req) {
    w.Write("Hello, World!")
}

mux.Handle("/", loggingMiddleware(handler))

ListenAndServe(":8080", mux)
```

## Comparison with Go

| Tau http.tau | Go net/http |
|-------------|-------------|
| `Get(url)` | `http.Get(url)` |
| `Post(url, ct, body)` | `http.Post(url, ct, body)` |
| `ListenAndServe(addr, nil)` | `http.ListenAndServe(addr, nil)` |
| `HandleFunc(pattern, fn)` | `http.HandleFunc(pattern, fn)` |
| `NewServeMux()` | `http.NewServeMux()` |
| `NewClient()` | `&http.Client{}` |
| `w.Write(data)` | `w.Write([]byte(data))` |
| `w.WriteHeader(code)` | `w.WriteHeader(code)` |
| `StatusOK` | `http.StatusOK` |

## Dependencies

- `stdlib/net.tau` - Network primitives
- `stdlib/strings.tau` - String manipulation

## Limitations

- HTTP/1.1 only (no HTTP/2 support)
- No HTTPS/TLS support
- No chunked transfer encoding
- No automatic redirect following
- No cookie handling
- No compression support
- Simplified header handling (case-sensitive)

## Future Enhancements

Potential additions to match more of Go's net/http:

- Cookie support (http.Cookie, http.SetCookie)
- Request/Response cloning
- Context support for cancellation
- Timeout configuration
- Keep-alive connections
- Request body streaming
- Multipart form data
- URL query parameter parsing
- More complete header handling

## Examples

See `examples/http_examples.tau` for comprehensive examples including:
1. Simple HTTP server
2. Multi-route server with ServeMux
3. HTTP client GET requests
4. Custom request headers
5. POST requests with JSON
6. REST API server
7. Static file server
8. Middleware patterns
9. Custom server configuration
10. Error handling and status codes

## License

Part of the Tau programming language standard library.
