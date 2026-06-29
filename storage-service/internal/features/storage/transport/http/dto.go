package transport_http

type UploadURLRequest struct {
	Filename string `json:"filename" binding:"required"`
}

type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
}
