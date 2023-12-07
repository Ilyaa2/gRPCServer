package domain

// AbsenceOptions used to decode the id of the reason into Emoji and Description
type AbsenceOptions struct {
	reasons map[int]Reason
}

type Reason struct {
	Description string
	Emoji       string
}

func NewAbsenceOptions() *AbsenceOptions {
	return initOptions()
}

func (a *AbsenceOptions) GetReason(i int) (Reason, bool) {
	r, ok := a.reasons[i]
	return r, ok
}

func initOptions() *AbsenceOptions {
	return &AbsenceOptions{
		reasons: map[int]Reason{
			1:  {"Личные дела", "(🏠)"},
			2:  {"Гостевой пропуск", ""},
			3:  {"Командировка", "(✈)"},
			4:  {"Местная командировка", "(✈)"},
			5:  {"Болезнь", "(🌡)"},
			6:  {"Больничный лист", "(🌡)"},
			7:  {"Ночные работы", ""},
			8:  {"Дежурство", ""},
			9:  {"Учеба", "(🎓)"},
			10: {"Удаленная работа", "(🏠)"},
			11: {"Отпуск", "(☀)"},
			12: {"Отпуск за свой счет", "(☀)"},
			13: {"Отсутствие с отработкой", "(☀)"},
		},
	}
}
