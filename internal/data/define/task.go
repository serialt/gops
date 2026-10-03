package define

type TaskJobListCond struct {
	PageCond
	Keyword  string
	TaskType string
	Status   string
}

type TaskTemplateListCond struct {
	PageCond
	Keyword  string
	Category string
	Platform string
	Status   *int64
}
