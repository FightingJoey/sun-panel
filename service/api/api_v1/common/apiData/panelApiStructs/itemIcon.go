package panelApiStructs

import (
	"sun-panel/models"
)

type ItemIconEditRequest struct {
	models.ItemIcon
	IconJson string
}

type ItemIconSaveSortItem struct {
	Id              uint `json:"id"`
	Sort            uint `json:"sort"`
	ItemIconGroupId uint `json:"itemIconGroupId"`
}

type ItemIconSaveSortRequest struct {
	SortItems       []ItemIconSaveSortItem `json:"sortItems"`
	ItemIconGroupId uint                   `json:"itemIconGroupId"` // kept for backward compatibility; ignored when sortItems carry group ids
}

type ItemIconGetSiteFaviconReq struct {
	Url string `json:"url"`
}

type ItemIconGetSiteFaviconResp struct {
	IconUrl string `json:"iconUrl"`
}
