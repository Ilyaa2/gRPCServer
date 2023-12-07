package domain

// JobsQueue - this is the composition of all job queues. Could be extended.
type JobsQueue struct {
	AbsenceJQ *chan AbsenceJob
}
