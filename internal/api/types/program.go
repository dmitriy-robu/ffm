package types

type Program struct {
	ID       int64
	Name     string
	GoalID   int64
	LevelID  int64
	PeriodID int64
}

type Level struct {
	ID   int64
	Name string
}

type Period struct {
	ID   int64
	Name string
}

type ProgramMonth struct {
	ID        int64
	ProgramID int64
	Month     int
}
