package domain

type EmployeeData struct {
	Status string              `json:"status"`
	Data   []EmployeeInnerData `json:"data"`
}

type EmployeeInnerData struct {
	Id          int    `json:"id"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	WorkPhone   string `json:"workPhone"`
}

type AbsenceReason struct {
	Status string              `json:"status"`
	Data   []AbsenceReasonData `json:"data"`
}

type AbsenceReasonData struct {
	Id          int    `json:"id"`
	PersonId    int    `json:"personId"`
	CreatedDate string `json:"createdDate"`
	DateFrom    string `json:"dateFrom"`
	DateTo      string `json:"dateTo"`
	ReasonId    int    `json:"reasonId"`
}
