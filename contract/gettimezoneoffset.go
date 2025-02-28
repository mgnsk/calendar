package contract

// GetTimezoneRequest is a request to fetch timezone by geo coordinates.
type GetTimezoneRequest struct {
	Latitude     string `query:"latitude"`
	Longitude    string `query:"longitude"`
	UserTimezone string `query:"user_timezone"`
}

// GetTimezoneResponse is the response to GetTimezoneRequest.
type GetTimezoneResponse struct {
	TimezoneOffset int `json:"timezone_offset"`
}
