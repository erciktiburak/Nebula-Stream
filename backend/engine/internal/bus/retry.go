package bus

type RetryPolicy struct {
	MaxAttempts       int
	DeadLetterSubject string
}
