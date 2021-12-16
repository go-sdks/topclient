package resp

type ResponseBase struct {
	Code          string `json:"code"`
	Msg           string `json:"msg"`
	SubCode       string `json:"sub_code"`
	SubMsg        string `json:"sub_msg"`
	RequestID     string `json:"request_id"`
	ErrorResponse *struct {
		Msg     string `json:"msg"`
		Code    int    `json:"code"`
		SubCode string `json:"sub_code"`
		SubMsg  string `json:"sub_msg"`
	} `json:"error_response"`
}
