package errors

import (
	"fmt"
)

// ErrorCode represents a specific error type for automation
type ErrorCode int

const (
	// General errors
	ErrUnknown ErrorCode = iota
	ErrInvalidInput
	ErrFileNotFound
	ErrPermissionDenied

	// Initialization errors
	ErrVaultNotInitialized
	ErrMachineNotInitialized
	ErrVaultAlreadyExists

	// Crypto errors
	ErrKeyGenerationFailed
	ErrEncryptionFailed
	ErrDecryptionFailed
	ErrInvalidKey
	ErrAccessDenied

	// Git errors
	ErrGitNotAvailable
	ErrGitOperationFailed
	ErrMergeConflict
	ErrRemoteUnreachable

	// Machine errors
	ErrMachineNotFound
	ErrMachineAlreadyExists
	ErrInvalidMachineID

	// Secret errors
	ErrSecretNotFound
	ErrEnvironmentNotFound
	ErrNoSecretsToEncrypt
	ErrInvalidSecretFormat
)

// NvoltError is a custom error type with context and recovery suggestions
type NvoltError struct {
	Code       ErrorCode
	Message    string
	Context    string
	Suggestion string
	Err        error
}

// Error implements the error interface
func (e *NvoltError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *NvoltError) Unwrap() error {
	return e.Err
}

// FullMessage returns a detailed error message with context and suggestions
func (e *NvoltError) FullMessage() string {
	msg := e.Error()

	if e.Context != "" {
		msg += fmt.Sprintf("\n\nContext: %s", e.Context)
	}

	if e.Suggestion != "" {
		msg += fmt.Sprintf("\n\nSuggestion: %s", e.Suggestion)
	}

	return msg
}

// New creates a new NvoltError
func New(code ErrorCode, message string) *NvoltError {
	return &NvoltError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *NvoltError {
	return &NvoltError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithContext adds context to an error
func (e *NvoltError) WithContext(context string) *NvoltError {
	e.Context = context
	return e
}

// WithSuggestion adds a recovery suggestion to an error
func (e *NvoltError) WithSuggestion(suggestion string) *NvoltError {
	e.Suggestion = suggestion
	return e
}

// Common error constructors for convenience

func NewVaultNotInitialized() *NvoltError {
	return New(ErrVaultNotInitialized, "vault not initialized").
		WithSuggestion("Run 'nvolt init' to initialize the vault, or 'nvolt init --repo <org/repo>' for global mode")
}

func NewMachineNotInitialized() *NvoltError {
	return New(ErrMachineNotInitialized, "machine not initialized").
		WithSuggestion("Run 'nvolt init' to create machine keypair")
}

func NewAccessDenied(environment string) *NvoltError {
	return New(ErrAccessDenied, fmt.Sprintf("access denied to '%s' environment", environment)).
		WithSuggestion("Request access from someone with push permissions, or use a machine that has access to this environment")
}

func NewEnvironmentNotFound(environment string) *NvoltError {
	return New(ErrEnvironmentNotFound, fmt.Sprintf("environment '%s' not found", environment)).
		WithSuggestion("Use 'nvolt push' to create secrets for this environment first")
}

func NewGitConflict() *NvoltError {
	return New(ErrMergeConflict, "git merge conflict detected").
		WithSuggestion("Resolve conflicts manually in the repository, then retry the operation")
}

func NewInvalidInput(field, value, reason string) *NvoltError {
	return New(ErrInvalidInput, fmt.Sprintf("invalid %s: %s", field, value)).
		WithContext(reason)
}

func NewFileNotFound(path string) *NvoltError {
	return New(ErrFileNotFound, fmt.Sprintf("file not found: %s", path))
}

func NewEncryptionFailed(err error) *NvoltError {
	return Wrap(err, ErrEncryptionFailed, "failed to encrypt secret").
		WithSuggestion("Check that the master key is valid and properly generated")
}

func NewDecryptionFailed(err error) *NvoltError {
	return Wrap(err, ErrDecryptionFailed, "failed to decrypt secret").
		WithSuggestion("Ensure you have access to this environment and the master key hasn't been rotated")
}
