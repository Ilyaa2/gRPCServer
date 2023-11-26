package domain

import (
	"bufio"
	"os"
	"path"
	"strconv"
	"strings"
)

type AbsenceOptions struct {
	reasons map[int]Reason
}

type Reason struct {
	Description string
	Emoji       string
}

// TODO исправить
const dirPath = "C:\\Users\\User\\GolandProjects\\gRPCServer\\internal\\domain\\static"

func NewAbsenceOptions(filename string) (*AbsenceOptions, error) {
	//todo
	return phonyParse(filename)
}

func (a *AbsenceOptions) GetReason(i int) (Reason, bool) {
	r, ok := a.reasons[i]
	return r, ok
}

func phonyParse(filename string) (*AbsenceOptions, error) {
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
	}, nil
}

// todo СДЕЛАТЬ ЭТО
// "7 Ночные работы" - некорретно //"8 Дежурство" - ошибка.
func parseAbsenceOptionsFile(filename string) (*AbsenceOptions, error) {
	m := map[int]Reason{}
	file, err := os.Open(path.Join(dirPath, filename))
	defer file.Close()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		position1 := strings.Index(str, " ")
		position2 := strings.LastIndex(str, " ")
		idx, err := strconv.Atoi(str[:position1])
		if err != nil {
			return nil, err
		}
		description := str[position1+1 : position2]
		emoji := str[position2:]
		m[idx] = Reason{
			Description: description,
			Emoji:       emoji,
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return &AbsenceOptions{reasons: m}, nil
}
