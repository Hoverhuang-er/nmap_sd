package sd

type RegisterTarget struct {
	Targets []string `json:"targets"`
	Labels  Labels   `json:"labels"`
}
type Labels struct {
	Env string `json:"env"`
	Job string `json:"job"`
}
type FileEncode struct {
	Encrypt bool  `json:"encrypt"`
	FileName string `json:"file_name"`
	FileBodyB64 string `json:"file_body_b64"`
	LargeFile bool `json:"large_file"`
	BodySize int64 `json:"body_size"`
}