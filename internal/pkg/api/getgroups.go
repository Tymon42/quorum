package api

import (
	"net/http"
	"sort"

	"github.com/rumsystem/quorum/internal/pkg/chain"
	"github.com/labstack/echo/v4"
)

type groupInfo struct {
	GroupId        string   `json:"group_id"`
	GroupName      string   `json:"group_name"`
	OwnerPubKey    string   `json:"owner_pubkey"`
	UserPubkey     string   `json:"user_pubkey"`
	ConsensusType  string   `json:"consensus_type"`
	EncryptionType string   `json:"encryption_type"`
	CipherKey      string   `json:"cipher_key"`
	AppKey         string   `json:"app_key"`
	LastUpdated    int64    `json:"last_updated"`
	HighestHeight  int64    `json:"highest_height"`
	HighestBlockId []string `json:"highest_block_id"`
	GroupStatus    string   `json:"group_status"`
}

type GroupInfoList struct {
	GroupInfos []*groupInfo `json:"groups"`
}

func (s *GroupInfoList) Len() int { return len(s.GroupInfos) }
func (s *GroupInfoList) Swap(i, j int) {
	s.GroupInfos[i], s.GroupInfos[j] = s.GroupInfos[j], s.GroupInfos[i]
}

func (s *GroupInfoList) Less(i, j int) bool {
	return s.GroupInfos[i].GroupName < s.GroupInfos[j].GroupName
}

// @Tags Groups
// @Summary GetGroups
// @Description Get all joined groups
// @Produce json
// @Success 200 {object} GroupInfoList
// @Router /api/v1/groups [get]
func (h *Handler) GetGroups(c echo.Context) (err error) {
	var groups []*groupInfo
	for _, value := range chain.GetNodeCtx().Groups {
		var group *groupInfo
		group = &groupInfo{}

		group.OwnerPubKey = value.Item.OwnerPubKey
		group.GroupId = value.Item.GroupId
		group.GroupName = value.Item.GroupName
		group.OwnerPubKey = value.Item.OwnerPubKey
		group.UserPubkey = value.Item.UserSignPubkey
		group.ConsensusType = value.Item.ConsenseType.String()
		group.EncryptionType = value.Item.EncryptType.String()
		group.CipherKey = value.Item.CipherKey
		group.AppKey = value.Item.AppKey
		group.LastUpdated = value.Item.LastUpdate
		group.HighestHeight = value.Item.HighestHeight
		group.HighestBlockId = value.Item.HighestBlockId

		switch value.ChainCtx.Syncer.Status {
		case chain.SYNCING_BACKWARD:
			group.GroupStatus = "SYNCING"
		case chain.SYNCING_FORWARD:
			group.GroupStatus = "SYNCING"
		case chain.SYNC_FAILED:
			group.GroupStatus = "SYNC_FAILED"
		case chain.IDLE:
			group.GroupStatus = "IDLE"
		}
		groups = append(groups, group)
	}

	ret := GroupInfoList{groups}
	sort.Sort(&ret)
	return c.JSON(http.StatusOK, &ret)
}
