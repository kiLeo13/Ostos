package families

import "testing"

func TestRolePolicyMatrix(t *testing.T) {
	actions := []Action{ActionView, ActionEditSettings, ActionInviteMember, ActionRemoveMember, ActionManageManager, ActionTransferOwnership, ActionDelete}
	expected := map[Role]map[Action]bool{
		RoleOwner:   {ActionView: true, ActionEditSettings: true, ActionInviteMember: true, ActionRemoveMember: true, ActionManageManager: true, ActionTransferOwnership: true, ActionDelete: true},
		RoleManager: {ActionView: true, ActionEditSettings: true, ActionInviteMember: true, ActionRemoveMember: true},
		RoleMember:  {ActionView: true},
	}
	for _, role := range []Role{RoleOwner, RoleManager, RoleMember} {
		for _, action := range actions {
			if got := Allows(role, action); got != expected[role][action] {
				t.Errorf("Allows(%q, %q) = %t", role, action, got)
			}
		}
	}
}
