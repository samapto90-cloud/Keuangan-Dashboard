package mmo

import (
	"os"
	"strings"
)

const (
	RolePlayer     = "PLAYER"
	RoleModerator  = "MODERATOR"
	RoleAdmin      = "ADMIN"
	RoleSuperAdmin = "SUPER_ADMIN"

	PermPlayerView      = "PLAYER_VIEW"
	PermPlayerEdit      = "PLAYER_EDIT"
	PermPlayerBan       = "PLAYER_BAN"
	PermQuestionView    = "QUESTION_VIEW"
	PermQuestionCreate  = "QUESTION_CREATE"
	PermQuestionEdit    = "QUESTION_EDIT"
	PermQuestionDelete  = "QUESTION_DELETE"
	PermMatchView       = "MATCH_VIEW"
	PermMatchTerminate  = "MATCH_TERMINATE"
	PermReportView      = "REPORT_VIEW"
	PermReportResolve   = "REPORT_RESOLVE"
	PermConfigView      = "CONFIG_VIEW"
	PermConfigEdit      = "CONFIG_EDIT"
	PermAdminManage     = "ADMIN_MANAGE"
	PermAchievementEdit = "ACHIEVEMENT_EDIT"
	PermRewardEdit      = "REWARD_EDIT"
	PermRankEdit        = "RANK_EDIT"
	PermSeasonManage    = "SEASON_MANAGE"
	PermAuditView       = "AUDIT_VIEW"
)

func NormalizeRole(role string) string {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case RoleModerator:
		return RoleModerator
	case RoleAdmin:
		return RoleAdmin
	case RoleSuperAdmin:
		return RoleSuperAdmin
	default:
		return RolePlayer
	}
}

func RolePerms(role string) map[string]bool {
	out := map[string]bool{}
	add := func(ps ...string) {
		for _, p := range ps {
			out[p] = true
		}
	}
	switch NormalizeRole(role) {
	case RoleModerator:
		add(PermPlayerView, PermMatchView, PermReportView, PermReportResolve, PermAuditView)
	case RoleAdmin:
		add(PermPlayerView, PermPlayerEdit, PermPlayerBan,
			PermQuestionView, PermQuestionCreate, PermQuestionEdit, PermQuestionDelete,
			PermMatchView, PermMatchTerminate, PermReportView, PermReportResolve,
			PermConfigView, PermAchievementEdit, PermRewardEdit, PermAuditView)
	case RoleSuperAdmin:
		add(PermPlayerView, PermPlayerEdit, PermPlayerBan,
			PermQuestionView, PermQuestionCreate, PermQuestionEdit, PermQuestionDelete,
			PermMatchView, PermMatchTerminate, PermReportView, PermReportResolve,
			PermConfigView, PermConfigEdit, PermAdminManage, PermAchievementEdit, PermRewardEdit,
			PermRankEdit, PermSeasonManage, PermAuditView)
	}
	return out
}

func HasPerm(role, perm string) bool {
	return RolePerms(role)[perm]
}

func bootstrapSuperUser() string {
	return strings.ToLower(strings.TrimSpace(os.Getenv("ULAR_SUPER_ADMIN")))
}
