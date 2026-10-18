package families

type Action string

const (
	ActionView              Action = "view"
	ActionEditSettings      Action = "edit_settings"
	ActionInviteMember      Action = "invite_member"
	ActionRemoveMember      Action = "remove_member"
	ActionManageManager     Action = "manage_manager"
	ActionTransferOwnership Action = "transfer_ownership"
	ActionDelete            Action = "delete"
)

func Allows(role Role, action Action) bool {
	switch action {
	case ActionView:
		return role == RoleOwner || role == RoleManager || role == RoleMember
	case ActionEditSettings, ActionInviteMember, ActionRemoveMember:
		return role == RoleOwner || role == RoleManager
	case ActionManageManager, ActionTransferOwnership, ActionDelete:
		return role == RoleOwner
	default:
		return false
	}
}
