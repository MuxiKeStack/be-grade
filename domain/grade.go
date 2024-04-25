package domain

type Grade struct {
	Id       int64
	CourseId int64
	Uid      int64
	Regular  float64
	Final    float64
	Total    float64
	Year     string
	Term     string
}
