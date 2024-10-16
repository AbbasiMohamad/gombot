package entities

type Application struct {
	ID            uint
	JobID         uint
	Name          string
	PersianName   string
	Status        ApplicationStatus
	GitlabUrl     string
	Branch        string
	NeedToApprove bool
}

type ApplicationStatus string // TODO: GPT this statement

const (
	Declared ApplicationStatus = "declared"
	Pending  ApplicationStatus = "pending"
	//Approved     ApplicationStatus = "approved"
	//InProgress   ApplicationStatus = "inprogress"
	BuildFailed ApplicationStatus = "buildfailed"
	//Built        ApplicationStatus = "built"
	DeployFailed ApplicationStatus = "deployfailed"
	Deployed     ApplicationStatus = "deployed"
	//Canceled     ApplicationStatus = "canceled"
)
