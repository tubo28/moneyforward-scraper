package mf

import "time"

type MFTransaction struct {
	TransactionID     string
	IsCalculateTarget bool
	Date              time.Time
	Content           string
	Amount            int
	Institution       string
	LargeCategory     string
	MiddleCategory    string
	Memo              string
	CollectedDate     time.Time
}
