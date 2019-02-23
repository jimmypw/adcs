package adcs

// WebEnrollmentRequest interface is implemented by the
// various kinds of
type WebEnrollmentRequest interface {
	Submit() (WebEnrollmentResponse, error)
	GetServer() *WebEnrollmentServer
}
