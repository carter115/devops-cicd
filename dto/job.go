package dto

type CreateJobOutput struct {
	Id string `json:"job_id"`
}

type GetJobOutput struct {
	Id        string      `json:"id"`
	Cost      string      `json:"cost"`
	Status    string      `json:"status"`
	PhaseList interface{} `json:"phase_list"`
}

type JobOutput struct {
	Id     string `json:"id"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Status string `json:"status"`
}

type ListJobOutput []JobOutput
