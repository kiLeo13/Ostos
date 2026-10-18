package request

type FamilyCreateRequest struct {
	Name string `json:"name" validate:"required,notblank,min=1,max=100"`
}

type FamilyUpdateRequest struct {
	Name string `json:"name" validate:"required,notblank,min=1,max=100"`
}

type MemberRoleUpdateRequest struct {
	Role string `json:"role" validate:"required,oneof=manager member"`
}

type TransferOwnershipRequest struct {
	NewOwnerID string `json:"new_owner_id" validate:"required,snowflake_id"`
}

type InvitationCreateRequest struct {
	Username *string `json:"username" validate:"omitempty,username,excluded_with=UserID,required_without=UserID"`
	UserID   *string `json:"user_id" validate:"omitempty,snowflake_id,excluded_with=Username,required_without=Username"`
}
