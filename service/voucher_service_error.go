package service

type VoucherServiceErrorKind string

const (
	VoucherServiceErrorValidation  VoucherServiceErrorKind = "validation"
	VoucherServiceErrorNotFound    VoucherServiceErrorKind = "not_found"
	VoucherServiceErrorConflict    VoucherServiceErrorKind = "conflict"
	VoucherServiceErrorUnavailable VoucherServiceErrorKind = "unavailable"
	VoucherServiceErrorInternal    VoucherServiceErrorKind = "internal"
)

type VoucherServiceError struct {
	Kind    VoucherServiceErrorKind
	Message string
}

func (err *VoucherServiceError) Error() string {
	return err.Message
}
