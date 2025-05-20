//go:build linux
package parser

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "net/textproto"
    "strconv"
    "strings"
    "github.com/stanleydv12/ginx/internal/entity"
)

const (
	HTTPMethodGet     = "GET"
	HTTPMethodPost    = "POST"
	HTTPMethodPut     = "PUT"
	HTTPMethodDelete  = "DELETE"
	HTTPMethodHead    = "HEAD"
	HTTPMethodOptions = "OPTIONS"
	HTTPMethodPatch   = "PATCH"

	HTTPProtocolHTTP11 = "HTTP/1.1"

	HTTPStatusCodeOK = 200
	HTTPStatusCodeNotFound = 404
	HTTPStatusCodeMethodNotAllowed = 405
	HTTPStatusCodeInternalServerError = 500

	HTTPStatusTextOK = "OK"
	HTTPStatusTextNotFound = "Not Found"
	HTTPStatusTextMethodNotAllowed = "Method Not Allowed"
	HTTPStatusTextInternalServerError = "Internal Server Error"
)

type HTTPParser struct{}

func NewHTTPParser() HTTPParser {
    return HTTPParser{}
}

func (p *HTTPParser) ParseHTTPRequest(data []byte) (*entity.HTTPRequest, error) {
    reader := bufio.NewReader(bytes.NewReader(data))
    
    // Read request line
    line, _, err := reader.ReadLine()
    if err != nil {
        return nil, err
    }
    
    // Parse request line
    parts := strings.SplitN(string(line), " ", 3)
    if len(parts) < 3 {
        return nil, fmt.Errorf("malformed HTTP request")
    }
    
    req := &entity.HTTPRequest{
        Method:   parts[0],
        Path:     parts[1],
        Protocol: parts[2],
        Headers:  make(map[string]string),
        Raw:      make([]byte, len(data)),
    }
    copy(req.Raw, data)
    
    // Parse headers
    tp := textproto.NewReader(bufio.NewReader(reader))
    for {
        header, err := tp.ReadLine()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        if header == "" {
            break // End of headers
        }
        
        // Split header into key-value
        parts := strings.SplitN(header, ":", 2)
        if len(parts) != 2 {
            continue
        }
        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        req.Headers[key] = value
    }
    
    // Read body if Content-Length is present
    if cl, ok := req.Headers["Content-Length"]; ok {
        contentLength := 0
        _, err := fmt.Sscanf(cl, "%d", &contentLength)
        if err == nil && contentLength > 0 {
            body := make([]byte, contentLength)
            _, err := io.ReadFull(reader, body)
            if err != nil {
                return nil, err
            }
            req.Body = body
        }
    }
    
    return req, nil
}

func (p *HTTPParser) ParseHTTPResponse(data []byte) (*entity.HTTPResponse, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	
	// Read status line
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read status line: %v", err)
	}
	statusLine = strings.TrimSpace(statusLine)

	// Parse status line (e.g., "HTTP/1.1 200 OK")
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("malformed status line: %s", statusLine)
	}

	// Parse status code
	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %v", err)
	}

	response := &entity.HTTPResponse{
		StatusCode: statusCode,
		Headers:   make(map[string]string),
		Raw:       make([]byte, len(data)),
	}
	copy(response.Raw, data)

	// Read headers
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			break
		}
			line = strings.TrimSpace(line)
		
		// End of headers
		if line == "" {
			break
		}

		// Parse header (e.g., "Content-Type: application/json")
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) != 2 {
			continue // Skip malformed headers
		}

		key := strings.TrimSpace(headerParts[0])
		value := strings.TrimSpace(headerParts[1])
		response.Headers[key] = value
	}

	// Read body if Content-Length is present
	if contentLength, ok := response.Headers["Content-Length"]; ok {
		length, err := strconv.Atoi(contentLength)
		if err == nil && length > 0 {
			body := make([]byte, length)
			n, err := reader.Read(body)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("failed to read response body: %v", err)
			}
			response.Body = body[:n]
		}
	} else {
		// If no Content-Length, read until EOF (for responses with Transfer-Encoding: chunked or connection close)
		body, err := io.ReadAll(reader)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		response.Body = body
	}

	return response, nil
}

func (p *HTTPParser) RebuildRequest(req *entity.HTTPRequest) []byte {
    var buf bytes.Buffer
    
    // Write request line
    buf.WriteString(fmt.Sprintf("%s %s %s\r\n", req.Method, req.Path, req.Protocol))
    
    // Write headers
    for key, value := range req.Headers {
        buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
    }
    
    // End of headers
    buf.WriteString("\r\n")
    
    // Write body if exists
    if len(req.Body) > 0 {
        buf.Write(req.Body)
    }
    
    return buf.Bytes()
}


func (p *HTTPParser) RebuildResponse(resp *entity.HTTPResponse) []byte {
    var buf bytes.Buffer
    
    // Write status line
    buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, HTTPStatusCode(resp.StatusCode)))
    
    // Write headers
    for key, value := range resp.Headers {
        buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
    }
    
    // End of headers
    buf.WriteString("\r\n")
    
    // Write body if exists
    if len(resp.Body) > 0 {
        buf.Write(resp.Body)
    }
    
    return buf.Bytes()
}

func HTTPStatusCode(statusCode int) string {
	switch statusCode {
	case HTTPStatusCodeOK:
		return HTTPStatusTextOK
	case HTTPStatusCodeNotFound:
		return HTTPStatusTextNotFound
	case HTTPStatusCodeMethodNotAllowed:
		return HTTPStatusTextMethodNotAllowed
	case HTTPStatusCodeInternalServerError:
		return HTTPStatusTextInternalServerError
	default:
		return fmt.Sprintf("Unknown status code: %d", statusCode)
	}
}
