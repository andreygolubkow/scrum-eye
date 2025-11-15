package domain

import "time"

type WorkItemType string

const (
	WorkItemStory   WorkItemType = "Story"
	WorkItemBug     WorkItemType = "Bug"
	WorkItemTask    WorkItemType = "Task"
	WorkItemEpic    WorkItemType = "Epic"
	WorkItemFeature WorkItemType = "Feature"
	WorkItemUnknown WorkItemType = "Unknown"
)

type WorkItem struct {
	ID   int
	Name string
	Type WorkItemType
}

type Sprint struct {
	ID        string
	Name      string
	StartDate *time.Time
	EndDate   *time.Time
	WorkItems []WorkItem
}
