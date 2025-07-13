# Security Guide

## Overview

This guide covers security considerations, best practices, and potential vulnerabilities when using the CLG logging package.

## Log Security Fundamentals

### Sensitive Data Protection

**Never log sensitive information:**
- Passwords or authentication tokens
- Social Security Numbers or personal IDs
- Credit card numbers or payment details
- Personal health information
- Cryptographic keys or secrets

```go
// ❌ BAD - Exposes sensitive data
logger.Info("User login: password=%s, ssn=%s", password, ssn)

// ✅ GOOD - Safe logging
logger.WithFields(map[string]interface{}{
    "user_id": userID,
    "login_attempt": true,
    "ip_address": getClientIP(),
}).Info("User login attempt")
```

### Data Sanitization

```go
import (
    "regexp"
    "strings"
)

// Sanitize sensitive data patterns
func sanitizeData(data string) string {
    // Remove credit card numbers
    ccRegex := regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
    data = ccRegex.ReplaceAllString(data, "****-****-****-****")
    
    // Remove SSN patterns
    ssnRegex := regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
    data = ssnRegex.ReplaceAllString(data, "***-**-****")
    
    // Remove email addresses if needed
    emailRegex := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
    data = emailRegex.ReplaceAllString(data, "****@****.***")
    
    return data
}

// Safe logging wrapper
func safeLog(logger *pim.Logger, level string, msg string, data interface{}) {
    sanitizedMsg := sanitizeData(fmt.Sprintf(msg, data))
    
    switch level {
    case "info":
        logger.Info(sanitizedMsg)
    case "error":
        logger.Error(sanitizedMsg)
    case "debug":
        logger.Debug(sanitizedMsg)
    }
}
```

## File Security

### File Permissions

```go
import "os"

// Set secure file permissions
func setupSecureLogger() *pim.Logger {
    // Create log directory with restricted permissions
    logDir := "/var/log/myapp"
    err := os.MkdirAll(logDir, 0750) // rwxr-x---
    if err != nil {
        panic(err)
    }
    
    logger := pim.NewLogger().
        EnableFileLogging().
        SetLogFile(filepath.Join(logDir, "app.log"))
    
    // Set file permissions after creation
    logFile := filepath.Join(logDir, "app.log")
    os.Chmod(logFile, 0640) // rw-r-----
    
    return logger
}
```

### Directory Security

```go
// Validate log file paths to prevent directory traversal
func validateLogPath(path string) error {
    // Clean the path
    cleanPath := filepath.Clean(path)
    
    // Check for directory traversal attempts
    if strings.Contains(cleanPath, "..") {
        return fmt.Errorf("invalid path: directory traversal detected")
    }
    
    // Ensure path is within allowed directory
    allowedDir := "/var/log/myapp"
    absPath, err := filepath.Abs(cleanPath)
    if err != nil {
        return err
    }
    
    absAllowed, err := filepath.Abs(allowedDir)
    if err != nil {
        return err
    }
    
    if !strings.HasPrefix(absPath, absAllowed) {
        return fmt.Errorf("path outside allowed directory")
    }
    
    return nil
}

// Safe log file creation
func createSecureLogger(logPath string) (*pim.Logger, error) {
    if err := validateLogPath(logPath); err != nil {
        return nil, err
    }
    
    logger := pim.NewLogger().
        EnableFileLogging().
        SetLogFile(logPath)
    
    return logger, nil
}
```

## Input Validation

### Log Injection Prevention

```go
import (
    "html"
    "strings"
    "unicode"
)

// Prevent log injection attacks
func sanitizeLogInput(input string) string {
    // Remove control characters
    cleaned := strings.Map(func(r rune) rune {
        if unicode.IsControl(r) && r != '\t' {
            return -1 // Remove character
        }
        return r
    }, input)
    
    // Escape HTML entities
    cleaned = html.EscapeString(cleaned)
    
    // Limit length to prevent DoS
    if len(cleaned) > 1000 {
        cleaned = cleaned[:1000] + "..."
    }
    
    return cleaned
}

// Safe user input logging
func logUserAction(logger *pim.Logger, userID string, action string, details string) {
    logger.WithFields(map[string]interface{}{
        "user_id": sanitizeLogInput(userID),
        "action":  sanitizeLogInput(action),
        "details": sanitizeLogInput(details),
    }).Info("User action performed")
}
```

### Format String Security

```go
// ❌ BAD - Vulnerable to format string attacks
func badLogging(logger *pim.Logger, userInput string) {
    logger.Info(userInput) // User could inject format specifiers
}

// ✅ GOOD - Safe format string usage
func goodLogging(logger *pim.Logger, userInput string) {
    logger.Info("User input: %s", userInput) // Safe format
    
    // Or use structured logging
    logger.WithFields(map[string]interface{}{
        "user_input": userInput,
    }).Info("User input received")
}
```

## Access Control

### Log File Access

```go
import (
    "os/user"
    "strconv"
    "syscall"
)

// Set specific user/group ownership
func setLogOwnership(logFile string, username string, groupname string) error {
    // Get user ID
    usr, err := user.Lookup(username)
    if err != nil {
        return err
    }
    uid, _ := strconv.Atoi(usr.Uid)
    
    // Get group ID
    grp, err := user.LookupGroup(groupname)
    if err != nil {
        return err
    }
    gid, _ := strconv.Atoi(grp.Gid)
    
    // Change ownership
    return syscall.Chown(logFile, uid, gid)
}

// Setup production logger with security
func setupProductionLogger() *pim.Logger {
    logFile := "/var/log/myapp/app.log"
    
    logger := pim.NewLogger().
        EnableFileLogging().
        SetLogFile(logFile)
    
    // Set secure permissions and ownership
    os.Chmod(logFile, 0640)
    setLogOwnership(logFile, "appuser", "loggroup")
    
    return logger
}
```

### Role-Based Logging

```go
type UserRole int

const (
    RoleUser UserRole = iota
    RoleAdmin
    RoleSystem
)

// Log different details based on user role
func logWithRole(logger *pim.Logger, role UserRole, action string, details map[string]interface{}) {
    baseFields := map[string]interface{}{
        "action": action,
        "role":   role.String(),
    }
    
    // Add role-specific details
    switch role {
    case RoleAdmin:
        // Admins get full details
        for k, v := range details {
            baseFields[k] = v
        }
    case RoleUser:
        // Users get limited details
        baseFields["user_action"] = true
    case RoleSystem:
        // System gets technical details
        baseFields["system_action"] = true
        baseFields["details"] = details
    }
    
    logger.WithFields(baseFields).Info("Action performed")
}
```

## Encryption and Privacy

### Log Encryption

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "io"
)

type EncryptedLogger struct {
    logger *pim.Logger
    gcm    cipher.AEAD
}

func NewEncryptedLogger(key []byte) (*EncryptedLogger, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    logger := pim.NewLogger().EnableFileLogging()
    
    return &EncryptedLogger{
        logger: logger,
        gcm:    gcm,
    }, nil
}

func (el *EncryptedLogger) encryptMessage(message string) (string, error) {
    nonce := make([]byte, el.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := el.gcm.Seal(nonce, nonce, []byte(message), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (el *EncryptedLogger) Info(msg string, args ...interface{}) {
    plaintext := fmt.Sprintf(msg, args...)
    encrypted, err := el.encryptMessage(plaintext)
    if err != nil {
        el.logger.Error("Failed to encrypt log message: %v", err)
        return
    }
    
    el.logger.Info("ENCRYPTED: %s", encrypted)
}
```

### PII Redaction

```go
import "regexp"

type PIIRedactor struct {
    patterns map[string]*regexp.Regexp
}

func NewPIIRedactor() *PIIRedactor {
    return &PIIRedactor{
        patterns: map[string]*regexp.Regexp{
            "email":    regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
            "phone":    regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b`),
            "ssn":      regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
            "ip":       regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
            "creditcard": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
        },
    }
}

func (p *PIIRedactor) Redact(text string) string {
    result := text
    
    for name, pattern := range p.patterns {
        switch name {
        case "email":
            result = pattern.ReplaceAllString(result, "[EMAIL_REDACTED]")
        case "phone":
            result = pattern.ReplaceAllString(result, "[PHONE_REDACTED]")
        case "ssn":
            result = pattern.ReplaceAllString(result, "[SSN_REDACTED]")
        case "ip":
            result = pattern.ReplaceAllString(result, "[IP_REDACTED]")
        case "creditcard":
            result = pattern.ReplaceAllString(result, "[CC_REDACTED]")
        }
    }
    
    return result
}

// Safe logger with PII redaction
type SafeLogger struct {
    logger   *pim.Logger
    redactor *PIIRedactor
}

func NewSafeLogger() *SafeLogger {
    return &SafeLogger{
        logger:   pim.NewLogger(),
        redactor: NewPIIRedactor(),
    }
}

func (sl *SafeLogger) Info(msg string, args ...interface{}) {
    message := fmt.Sprintf(msg, args...)
    redacted := sl.redactor.Redact(message)
    sl.logger.Info(redacted)
}
```

## Audit Logging

### Tamper-Evident Logging

```go
import (
    "crypto/sha256"
    "encoding/hex"
    "time"
)

type AuditLogger struct {
    logger     *pim.Logger
    lastHash   string
    sequence   uint64
}

func NewAuditLogger() *AuditLogger {
    return &AuditLogger{
        logger:   pim.NewLogger().EnableFileLogging(),
        lastHash: "0",
        sequence: 0,
    }
}

func (al *AuditLogger) LogAuditEvent(event string, details map[string]interface{}) {
    al.sequence++
    
    // Create audit record
    record := map[string]interface{}{
        "sequence":    al.sequence,
        "timestamp":   time.Now().Unix(),
        "event":       event,
        "details":     details,
        "prev_hash":   al.lastHash,
    }
    
    // Calculate hash
    recordStr := fmt.Sprintf("%v", record)
    hash := sha256.Sum256([]byte(recordStr))
    al.lastHash = hex.EncodeToString(hash[:])
    
    record["hash"] = al.lastHash
    
    al.logger.WithFields(record).Info("AUDIT")
}

// Verify audit log integrity
func (al *AuditLogger) VerifyIntegrity(logFile string) error {
    // Read audit log and verify hash chain
    // Implementation depends on your specific requirements
    return nil
}
```

### Compliance Logging

```go
// GDPR-compliant logging
type GDPRLogger struct {
    logger      *pim.Logger
    dataSubject string
}

func NewGDPRLogger(dataSubject string) *GDPRLogger {
    return &GDPRLogger{
        logger:      pim.NewLogger(),
        dataSubject: dataSubject,
    }
}

func (gl *GDPRLogger) LogProcessing(purpose string, legalBasis string, data map[string]interface{}) {
    gl.logger.WithFields(map[string]interface{}{
        "data_subject":    gl.dataSubject,
        "processing_purpose": purpose,
        "legal_basis":     legalBasis,
        "timestamp":       time.Now().Unix(),
        "data_categories": extractCategories(data),
    }).Info("GDPR_PROCESSING")
}

func extractCategories(data map[string]interface{}) []string {
    // Classify data types for GDPR compliance
    categories := []string{}
    
    for key := range data {
        switch key {
        case "email", "phone":
            categories = append(categories, "contact_data")
        case "name", "address":
            categories = append(categories, "identification_data")
        case "ip_address":
            categories = append(categories, "technical_data")
        }
    }
    
    return categories
}
```

## Security Monitoring

### Anomaly Detection

```go
import (
    "sync"
    "time"
)

type SecurityMonitor struct {
    logger      *pim.Logger
    rateLimiter map[string]*RateLimit
    mu          sync.RWMutex
}

type RateLimit struct {
    count     int
    window    time.Time
    threshold int
}

func NewSecurityMonitor() *SecurityMonitor {
    return &SecurityMonitor{
        logger:      pim.NewLogger().EnableFileLogging(),
        rateLimiter: make(map[string]*RateLimit),
    }
}

func (sm *SecurityMonitor) LogWithRateLimit(source string, event string) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    now := time.Now()
    rl, exists := sm.rateLimiter[source]
    
    if !exists {
        rl = &RateLimit{count: 0, window: now, threshold: 100}
        sm.rateLimiter[source] = rl
    }
    
    // Reset window if needed
    if now.Sub(rl.window) > time.Minute {
        rl.count = 0
        rl.window = now
    }
    
    rl.count++
    
    // Check for anomaly
    if rl.count > rl.threshold {
        sm.logger.WithFields(map[string]interface{}{
            "source":    source,
            "event":     event,
            "count":     rl.count,
            "threshold": rl.threshold,
            "severity":  "HIGH",
        }).Error("SECURITY_ANOMALY")
    } else {
        sm.logger.WithFields(map[string]interface{}{
            "source": source,
            "event":  event,
        }).Info("SECURITY_EVENT")
    }
}
```

## Best Practices Checklist

### Development
- [ ] Never log passwords, tokens, or secrets
- [ ] Sanitize user input before logging
- [ ] Use structured logging for better parsing
- [ ] Implement proper error handling
- [ ] Validate log file paths

### Production
- [ ] Set secure file permissions (640 or 644)
- [ ] Use dedicated log directories
- [ ] Implement log rotation
- [ ] Monitor disk space usage
- [ ] Set up log forwarding to SIEM

### Compliance
- [ ] Implement audit logging for sensitive operations
- [ ] Add retention policies
- [ ] Encrypt logs containing PII
- [ ] Document data processing activities
- [ ] Test incident response procedures

### Monitoring
- [ ] Set up anomaly detection
- [ ] Monitor for suspicious patterns
- [ ] Implement alerting for security events
- [ ] Regular security reviews
- [ ] Penetration testing

## Related Documentation

- [Configuration Guide](./CONFIGURATION.md)
- [Performance Guide](./PERFORMANCE.md)
- [Troubleshooting](./TROUBLESHOOTING.md)
- [API Reference](./api_reference.md)
