package domain

type WorkItemType string

const (
	WorkItemStory   WorkItemType = "Story"
	WorkItemBug     WorkItemType = "Bug"
	WorkItemTask    WorkItemType = "Task"
	WorkItemEpic    WorkItemType = "Epic"
	WorkItemFeature WorkItemType = "Feature"
)

type WorkItem struct {
	ID   string
	Name string
	Type WorkItemType
}

type Sprint struct {
	ID        string
	Name      string
	WorkItems []WorkItem
}
