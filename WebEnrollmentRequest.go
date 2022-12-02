package adcs

// WebEnrollmentRequest interface is implemented by the
// various kinds of requests that can be performed
type WebEnrollmentRequest interface {
	Submit() (WebEnrollmentResponse, error)
	GetServer() *WebEnrollmentServer
}
