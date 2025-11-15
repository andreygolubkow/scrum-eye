package collector

import (
	"scrum-eye/internal/domain"
	"scrum-eye/internal/sources/azureboards"
	"strings"
)

func MapODataWorkItems(src []azureboards.ODataWorkItem) []domain.WorkItem {
	dst := make([]domain.WorkItem, 0, len(src))

	for _, v := range src {
		wi := domain.WorkItem{
			ID:   v.ID,
			Name: v.Title,
			Type: normalizeWorkItemType(v.WorkItemType),
		}

		dst = append(dst, wi)
	}

	return dst
}

func normalizeWorkItemType(t string) domain.WorkItemType {
	switch strings.ToLower(t) {
	case "user story":
		return domain.WorkItemStory
	case "bug":
		return domain.WorkItemBug
	case "task":
		return domain.WorkItemTask
	case "epic":
		return domain.WorkItemEpic
	case "feature":
		return domain.WorkItemFeature
	default:
		return domain.WorkItemUnknown
	}
}
