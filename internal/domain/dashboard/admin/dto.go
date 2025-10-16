package admin

type AdminTotals struct {
	TotalUsers     int64 `json:"total_users"`
	TotalStudents  int64 `json:"total_students"`
	TotalTeachers  int64 `json:"total_teachers"`
	TotalCompanies int64 `json:"total_companies"`
}

type StudentStatistics struct {
	TotalTKJ  int64 `json:"total_tkj"`
	TotalTJA  int64 `json:"total_tja"`
	TotalPPLG int64 `json:"total_pplg"`
	TotalRPL  int64 `json:"total_rpl"`
}

type AdminDashboardData struct {
	Totals       AdminTotals       `json:"totals"`
	StudentStats StudentStatistics `json:"student_stats"`
}
