package enum

type ScheduleStatusEnum string

const (
	UPCOMMING ScheduleStatusEnum = "UPCOMMING"
	OVERDUE   ScheduleStatusEnum = "OVERDUE"
	CANCELED  ScheduleStatusEnum = "CANCELED"
	COMPLETED ScheduleStatusEnum = "COMPLETED"
)

func (e ScheduleStatusEnum) ToString() string {
	switch e {
	case UPCOMMING:
		return "UPCOMMING"
	case OVERDUE:
		return "OVERDUE"
	case CANCELED:
		return "CANCELED"
	case COMPLETED:
		return "COMPLETED"
	}
	return ""
}

func (e ScheduleStatusEnum) IsValid() bool {
	switch e {
	case UPCOMMING, OVERDUE, CANCELED, COMPLETED:
		return true
	}
	return false
}
