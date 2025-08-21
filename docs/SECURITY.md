# Security Improvements - Go Echo LiveView

## Overview
This document describes the security measures implemented in the Go Echo LiveView framework to protect against common web vulnerabilities.

## Security Features Implemented

### 1. WebSocket Message Validation (SEC-002) ✅

**File:** `liveview/security.go`, `liveview/page_content.go`

All incoming WebSocket messages are now validated before processing:

- **Message Size Limits**: Maximum 1MB per message
- **JSON Structure Validation**: Ensures valid JSON format
- **Type Validation**: Only allows predefined message types
- **ID Validation**: Prevents injection through ID fields
- **Event Name Validation**: Only alphanumeric event names allowed
- **Field Count Limits**: Maximum 100 fields per message
- **Rate Limiting**: 100 messages per minute per client

**Protection Against:**
- Buffer overflow attacks
- JSON injection
- DoS attacks through large messages
- Malformed data attacks

### 2. Template Sanitization (SEC-003) ✅

**File:** `liveview/security.go`, `liveview/model.go`

All templates are sanitized before rendering:

- **Script Tag Removal**: Removes all `<script>` tags
- **Event Handler Filtering**: Removes dangerous event handlers (preserves `send_event`)
- **IFrame Blocking**: Removes `<iframe>` elements
- **Object/Embed Removal**: Blocks `<object>` and `<embed>` tags
- **Template Size Limits**: Maximum 500KB per template
- **Go Template Preservation**: Maintains Go template syntax while sanitizing HTML

**Protection Against:**
- XSS (Cross-Site Scripting)
- HTML injection
- Template injection attacks

### 3. Path Traversal Prevention (SEC-004) ✅

**File:** `liveview/security.go`, `liveview/utils.go`

File operations are protected against path traversal:

- **Path Validation**: Blocks `..` in paths
- **Restricted Paths**: Blocks access to system directories (`/etc`, `/proc`, `/sys`, `/dev`)
- **Path Depth Limits**: Maximum 10 directory levels
- **Dangerous Character Filtering**: Blocks shell metacharacters
- **Clean Path Resolution**: Uses `filepath.Clean()` for path normalization

**Protection Against:**
- Directory traversal attacks
- Unauthorized file access
- System file exposure

### 4. Message Size Limits & Rate Limiting (SEC-005) ✅

**File:** `liveview/security.go`, `liveview/page_content.go`

Comprehensive limits to prevent resource exhaustion:

- **WebSocket Message Size**: 1MB maximum
- **Template Size**: 500KB maximum
- **Read/Write Buffer Size**: 1KB for WebSocket
- **Rate Limiting**: Per-client message throttling
- **Connection Limits**: Configurable per deployment

**Protection Against:**
- DoS attacks
- Memory exhaustion
- CPU exhaustion
- Bandwidth abuse

## Security Configuration

### Default Security Settings

```go
// In liveview/security.go
const (
    MaxMessageSize     = 1024 * 1024  // 1MB
    MaxTemplateSize    = 500 * 1024   // 500KB
    MaxPathDepth       = 10
    MaxEventNameLength = 100
    MaxFieldCount      = 100
)
```

### Custom Configuration

You can customize security settings using:

```go
import "github.com/arturoeanton/go-echo-live-view/liveview"

config := &liveview.SecurityConfig{
    MaxMessageSize:     2 * 1024 * 1024,  // 2MB
    MaxTemplateSize:    1024 * 1024,      // 1MB
    EnableSanitization: true,
    RestrictedPaths:    []string{"/etc", "/var", "/usr"},
}

liveview.SetSecurityConfig(config)
```

## Security Best Practices

### For Developers Using This Framework

1. **Never Trust User Input**
   - Always validate data received from clients
   - Use the provided validation functions

2. **Limit File Access**
   - Only allow access to necessary directories
   - Use the validated file functions

3. **Monitor Rate Limits**
   - Adjust rate limits based on your application needs
   - Monitor for abuse patterns

4. **Template Security**
   - Avoid using `EvalScript` in production
   - Sanitize all dynamic content in templates

5. **Authentication & Authorization**
   - Implement authentication middleware
   - Validate user permissions for sensitive operations

### Example: Secure Component

```go
type SecureComponent struct {
    *liveview.ComponentDriver[*SecureComponent]
    Data string
}

func (c *SecureComponent) GetTemplate() string {
    // Template will be automatically sanitized
    return `<div id="{{.IdComponent}}">
        {{.Data}} <!-- HTML will be escaped -->
    </div>`
}

func (c *SecureComponent) HandleEvent(data interface{}) {
    // Data is already validated by the framework
    // Additional application-specific validation here
    if validated := c.validateAppData(data); validated {
        c.Data = liveview.SanitizeHTML(data.(string))
        c.Commit()
    }
}
```

## Security Checklist

Before deploying to production:

- [ ] Enable all security features
- [ ] Configure appropriate rate limits
- [ ] Implement authentication
- [ ] Set up HTTPS/WSS
- [ ] Configure CORS properly
- [ ] Review template sanitization settings
- [ ] Test with security scanning tools
- [ ] Monitor for suspicious activity
- [ ] Regular security updates

## Reporting Security Issues

If you discover a security vulnerability, please:

1. **DO NOT** create a public GitHub issue
2. Email security details to the maintainers
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

## Future Security Enhancements

Planned improvements:

- [ ] Content Security Policy (CSP) headers
- [ ] CSRF token validation
- [ ] Session encryption
- [ ] OAuth2/OIDC integration
- [ ] Audit logging
- [ ] Security headers middleware
- [ ] Automated security testing

## Compliance

This framework implements security measures aligned with:

- OWASP Top 10 Web Application Security Risks
- CWE (Common Weakness Enumeration) guidelines
- Security best practices for WebSocket applications

## Version History

- **v1.1.0** - Added comprehensive security layer
  - WebSocket validation
  - Template sanitization
  - Path traversal prevention
  - Rate limiting
  - Message size limits

## License

Security improvements are part of the main project license.