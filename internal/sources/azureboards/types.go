package azureboards

import "time"

type Iteration struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Path       string              `json:"path"`
	Attributes IterationAttributes `json:"attributes"`
}

type IterationAttributes struct {
	StartDate  *time.Time `json:"startDate,omitempty"`
	FinishDate *time.Time `json:"finishDate,omitempty"`
	TimeFrame  string     `json:"timeFrame,omitempty"`
}

type iterationsListResponse struct {
	Value []Iteration `json:"value"`
	Count int         `json:"count"`
}

type iterationWorkItemsResponse struct {
	Count             int                `json:"count"`
	WorkItemRelations []WorkItemRelation `json:"workItemRelations"`
}

type WorkItemRelation struct {
	Rel    string       `json:"rel"`
	Source *WorkItemRef `json:"source,omitempty"`
	Target *WorkItemRef `json:"target,omitempty"`
}

type WorkItemRef struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type WorkItem struct {
	ID     int                    `json:"id"`
	Rev    int                    `json:"rev"`
	Fields map[string]interface{} `json:"fields"`
	URL    string                 `json:"url"`
}

type workItemsListResponse struct {
	Count int        `json:"count"`
	Value []WorkItem `json:"value"`
}
