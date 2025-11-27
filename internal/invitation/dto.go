package invitation

// InviteMembersRequest represents the request to invite members to a room (V2)
type InviteMembersRequest struct {
	RoomID    int64   `json:"roomId" binding:"required,gt=0"`
	MemberIDs []int64 `json:"memberIds" binding:"required,min=1,dive,gt=0"`
}

// InvitationInfoScrollResponse represents the response of invitation list
type InvitationInfoScrollResponse struct {
	Invitations []InvitationInfo `json:"invitations"`
}

//// InvitationStatusUpdateRequest represents the request to update invitation status
//type InvitationStatusUpdateRequest struct {
//	Status string `json:"status" binding:"required,oneof=ACCEPTED REJECTED"`
//}
