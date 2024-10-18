package dtos

type CreateApplicationDto struct {
	Name            string
	PersianName     string
	GitlabUrl       string
	GitlabProjectID int
	Branch          string
	NeedToApprove   bool
}
