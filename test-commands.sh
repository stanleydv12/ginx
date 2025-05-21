# Basic requests
curl -v http://localhost:8080/get
curl -v -H "Host: httpbin1" http://localhost:8080/get
curl -v -H "X-Target-Service: httpbin1" http://localhost:8080/get

# Different HTTP methods
curl -v -X POST -d "test=data" -H "Content-Type: application/x-www-form-urlencoded" http://localhost:8080/post
curl -v -X PUT -d "test=putdata" -H "Content-Type: application/json" http://localhost:8080/put
curl -v -X DELETE http://localhost:8080/delete

# Header testing
curl -v -H "X-Custom-Header: test" http://localhost:8080/headers
curl -v -H "X-Request-Id: 12345" -H "X-Forwarded-For: 192.168.1.1" http://localhost:8080/headers

# Status codes and delays
curl -v http://localhost:8080/status/201
curl -v http://localhost:8080/delay/1

# Different content types
curl -v -H "Accept: application/json" http://localhost:8080/json
curl -v -H "Accept: application/xml" http://localhost:8080/xml

# Authentication and redirects
curl -v -u "user:pass" http://localhost:8080/basic-auth/user/pass
curl -v -L http://localhost:8080/redirect-to?url=http://example.com