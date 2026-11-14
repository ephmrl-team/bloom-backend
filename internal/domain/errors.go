package domain

type AppError struct {
	Status int
	Code   string
	Msg    string
}

func (e *AppError) Error() string {
	return e.Msg
}

func NotFound(msg string) *AppError { return &AppError{Status: 404, Code: "not_found", Msg: msg} }
func AlreadyExists(msg string) *AppError {
	return &AppError{Status: 409, Code: "already_exists", Msg: msg}
}
func Failed(msg string) *AppError { return &AppError{Status: 500, Code: "failed", Msg: msg} }
func InternalError(msg string) *AppError {
	return &AppError{Status: 500, Code: "internal_error", Msg: msg}
}
func InvalidData(msg string) *AppError { return &AppError{Status: 400, Code: "invalid_data", Msg: msg} }
func InvalidRequest(msg string) *AppError {
	return &AppError{Status: 400, Code: "invalid_request", Msg: msg}
}
func UserNotFound(msg string) *AppError {
	return &AppError{Status: 400, Code: "user_not_found", Msg: msg}
}
func NoReceiverCipherText(msg string) *AppError {
	return &AppError{Status: 400, Code: "no_receiver_cipher_text", Msg: msg}
}
func NoSenderCipherText(msg string) *AppError {
	return &AppError{Status: 400, Code: "no_sender_cipher_text", Msg: msg}
}
func NoSyncKeyCiphertext(msg string) *AppError {
	return &AppError{Status: 400, Code: "no_sync_key_ciphertext", Msg: msg}
}
func NoSyncKeyNonce(msg string) *AppError {
	return &AppError{Status: 400, Code: "no_sync_key_nonce", Msg: msg}
}
func InvalidReceiverCiphertextLength(msg string) *AppError {
	return &AppError{Status: 400, Code: "invalid_receiver_ciphertext_length", Msg: msg}
}
func InvalidSenderCiphertextLength(msg string) *AppError {
	return &AppError{Status: 400, Code: "invalid_sender_ciphertext_length", Msg: msg}
}
func InvalidNonceLength(msg string) *AppError {
	return &AppError{Status: 400, Code: "invalid_nonce_length", Msg: msg}
}
func InvalidSyncKeyLength(msg string) *AppError {
	return &AppError{Status: 400, Code: "invalid_sync_key_length", Msg: msg}
}
func Expired(msg string) *AppError    { return &AppError{Status: 400, Code: "expired", Msg: msg} }
func NotExpired(msg string) *AppError { return &AppError{Status: 400, Code: "not_expired", Msg: msg} }
func Unauthorized(msg string) *AppError {
	return &AppError{Status: 401, Code: "unauthorized", Msg: msg}
}
