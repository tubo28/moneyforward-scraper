package mf

type MFTransaction struct {
	UserID            string
	TransactionID     string
	IsCalculateTarget bool
	DateRFC3339       string
	DateUnix          int64
	Content           string
	Amount            int
	Institution       string
	LargeCategory     string
	MiddleCategory    string
	Memo              string
}
