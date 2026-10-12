package panel

import (
	"math"
	"strings"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/models"
	"sun-panel/models/datatype"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type ItemIconGroup struct {
}

func normalizeNetworkConfig(cfg *datatype.NetworkConfig) {
	if cfg == nil {
		return
	}
	if len(cfg.NetworkStates) == 0 {
		*cfg = datatype.DefaultNetworkConfig()
		return
	}
	for i := range cfg.NetworkStates {
		cfg.NetworkStates[i].Host = strings.TrimSpace(cfg.NetworkStates[i].Host)
		cfg.NetworkStates[i].Name = strings.TrimSpace(cfg.NetworkStates[i].Name)
		cfg.NetworkStates[i].Type = strings.TrimSpace(cfg.NetworkStates[i].Type)
		cfg.NetworkStates[i].Id = strings.TrimSpace(cfg.NetworkStates[i].Id)
	}
	if cfg.ActiveNetworkStateId == "" {
		cfg.ActiveNetworkStateId = cfg.NetworkStates[0].Id
	}
}

func (a *ItemIconGroup) Edit(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := models.ItemIconGroup{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	req.UserId = userInfo.ID
	normalizeNetworkConfig(&req.NetworkConfig)

	if req.ID != 0 {
		// 修改 — use map update so custom Valuer NetworkConfig always persists
		updates := map[string]interface{}{
			"icon":           req.Icon,
			"title":          req.Title,
			"description":    req.Description,
			"user_id":        req.UserId,
			"network_config": req.NetworkConfig,
		}
		if req.Sort != 0 {
			updates["sort"] = req.Sort
		}
		if err := global.Db.Model(&models.ItemIconGroup{}).
			Where("id=? AND user_id=?", req.ID, userInfo.ID).
			Updates(updates).Error; err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return
		}
	} else {
		// 创建
		if err := global.Db.Create(&req).Error; err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return
		}
	}

	apiReturn.SuccessData(c, req)
}

type switchNetworkStateReq struct {
	Id                   uint   `json:"id"`
	ActiveNetworkStateId string `json:"activeNetworkStateId"`
}

// SwitchNetworkState only flips activeNetworkStateId using DB-stored hosts.
// Avoids the home page overwriting saved hosts with a stale in-memory config.
func (a *ItemIconGroup) SwitchNetworkState(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := switchNetworkStateReq{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	if req.Id == 0 || strings.TrimSpace(req.ActiveNetworkStateId) == "" {
		apiReturn.ErrorParamFomat(c, "id and activeNetworkStateId are required")
		return
	}

	var group models.ItemIconGroup
	if err := global.Db.Where("id=? AND user_id=?", req.Id, userInfo.ID).First(&group).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	normalizeNetworkConfig(&group.NetworkConfig)

	found := false
	for _, s := range group.NetworkConfig.NetworkStates {
		if s.Id == req.ActiveNetworkStateId {
			found = true
			break
		}
	}
	if !found {
		apiReturn.ErrorParamFomat(c, "network state not found")
		return
	}

	group.NetworkConfig.ActiveNetworkStateId = req.ActiveNetworkStateId
	if err := global.Db.Model(&models.ItemIconGroup{}).
		Where("id=? AND user_id=?", req.Id, userInfo.ID).
		Update("network_config", group.NetworkConfig).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.SuccessData(c, group)
}

func (a *ItemIconGroup) GetList(c *gin.Context) {

	userInfo, _ := base.GetCurrentUserInfo(c)
	groups := []models.ItemIconGroup{}

	err := global.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Order("sort ,created_at").Where("user_id=?", userInfo.ID).Find(&groups).Error; err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return err
		}

		// 判断分组是否为空，为空将自动创建默认分组
		if len(groups) == 0 {
			defaultGroup := models.ItemIconGroup{
				Title:         "APP",
				UserId:        userInfo.ID,
				Icon:          "material-symbols:ad-group-outline",
				NetworkConfig: datatype.DefaultNetworkConfig(),
			}
			if err := tx.Create(&defaultGroup).Error; err != nil {
				apiReturn.ErrorDatabase(c, err.Error())
				return err
			}

			// 并将当前账号下所有无分组的图标更新到当前组
			if err := tx.Model(&models.ItemIcon{}).Where("user_id=?", userInfo.ID).Update("item_icon_group_id", defaultGroup.ID).Error; err != nil {
				apiReturn.ErrorDatabase(c, err.Error())
				return err
			}

			groups = append(groups, defaultGroup)
		}

		// Ensure existing groups without network config get defaults and persist once
		for i := range groups {
			if len(groups[i].NetworkConfig.NetworkStates) == 0 {
				groups[i].NetworkConfig = datatype.DefaultNetworkConfig()
				_ = tx.Model(&models.ItemIconGroup{}).Where("id=?", groups[i].ID).
					Update("network_config", groups[i].NetworkConfig).Error
			}
		}

		// 返回 nil 提交事务
		return nil
	})

	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	} else {
		apiReturn.SuccessListData(c, groups, 0)
	}
}

func (a *ItemIconGroup) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	userInfo, _ := base.GetCurrentUserInfo(c)

	var count int64
	if err := global.Db.Model(&models.ItemIconGroup{}).Where(" user_id=?", userInfo.ID).Count(&count).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	} else {
		if math.Abs(float64(len(req.Ids))-float64(count)) < 1 {
			apiReturn.ErrorCode(c, 1201, "At least one must be retained", nil)
			return
		}

	}

	txErr := global.Db.Transaction(func(tx *gorm.DB) error {
		mitemIcon := models.ItemIcon{}
		if err := tx.Delete(&models.ItemIconGroup{}, "id in ? AND user_id=?", req.Ids, userInfo.ID).Error; err != nil {
			return err
		}

		if err := mitemIcon.DeleteByItemIconGroupIds(tx, userInfo.ID, req.Ids); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		apiReturn.ErrorDatabase(c, txErr.Error())
		return
	}

	apiReturn.Success(c)
}

// 保存排序
func (a *ItemIconGroup) SaveSort(c *gin.Context) {
	req := commonApiStructs.SortRequest{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)

	transactionErr := global.Db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		for _, v := range req.SortItems {
			if err := tx.Model(&models.ItemIconGroup{}).Where("user_id=? AND id=?", userInfo.ID, v.Id).Update("sort", v.Sort).Error; err != nil {
				// 返回任何错误都会回滚事务
				return err
			}
		}

		// 返回 nil 提交事务
		return nil
	})

	if transactionErr != nil {
		apiReturn.ErrorDatabase(c, transactionErr.Error())
		return
	}

	apiReturn.Success(c)
}
