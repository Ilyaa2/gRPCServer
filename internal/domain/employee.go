package domain

type EmployeeData struct {
	Status string              `json:"status"`
	Data   []EmployeeInnerData `json:"data"`
}

type EmployeeInnerData struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	WorkPhone   string `json:"workPhone"`
	Id          int    `json:"id"`
}

type AbsenceReason struct {
	Status string              `json:"status"`
	Data   []AbsenceReasonData `json:"data"`
}

type AbsenceReasonData struct {
	CreatedDate string `json:"createdDate"`
	DateFrom    string `json:"dateFrom"`
	DateTo      string `json:"dateTo"`
	Id          int    `json:"id"`
	ReasonId    int    `json:"reasonId"`
	PersonId    int    `json:"personId"`
}
