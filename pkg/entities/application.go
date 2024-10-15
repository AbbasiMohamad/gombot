package entities

type Application struct {
	Name      string
	Status    ApplicationStatus
	GitlabUrl string
	Branch    string
}

type ApplicationStatus string

const (
	Declared     ApplicationStatus = "declared"
	Pending      ApplicationStatus = "pending"
	Approved     ApplicationStatus = "approved"
	InProgress   ApplicationStatus = "inprogress"
	BuildFailed  ApplicationStatus = "buildfailed"
	Built        ApplicationStatus = "built"
	DeployFailed ApplicationStatus = "deployfailed"
	Deployed     ApplicationStatus = "deployed"
	Canceled     ApplicationStatus = "canceled"
)
