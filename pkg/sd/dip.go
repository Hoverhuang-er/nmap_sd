package sd

type RegisterTarget struct {
	Targets []string `json:"targets"`
	Labels  Labels   `json:"labels"`
}
type Labels struct {
	Env string `json:"env"`
	Job string `json:"job"`
}
