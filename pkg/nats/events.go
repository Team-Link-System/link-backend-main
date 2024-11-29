package nats

// ! 이벤트 정의 패키지
const (
	UserCreatedEvent   = "user.created"
	UserUpdatedEvent   = "user.updated"
	UserDeletedEvent   = "user.deleted"
	UserLoginEvent     = "user.login"
	UserLogoutEvent    = "user.logout"
	UserMentionedEvent = "user.mentioned"

	CompanyCreatedEvent = "company.created"
	CompanyUpdatedEvent = "company.updated"
	CompanyDeletedEvent = "company.deleted"

	RoleCreatedEvent = "role.created"
	RoleUpdatedEvent = "role.updated"
	RoleDeletedEvent = "role.deleted"

	//채팅방 참가
	ChatRoomJoinedEvent = "chat_room.joined"
	ChatRoomLeftEvent   = "chat_room.left"
	//채팅 메시지
	ChatMessageSentEvent = "chat_message.sent"

	//채팅방 관련
	ChatRoomCreatedEvent = "chat_room.created"
	ChatRoomUpdatedEvent = "chat_room.updated"
	ChatRoomDeletedEvent = "chat_room.deleted"

	//게시물 관련
	PostCreatedEvent = "post.created"
	PostUpdatedEvent = "post.updated"
	PostDeletedEvent = "post.deleted"

	//댓글 관련
	CommentCreatedEvent = "comment.created"
	CommentUpdatedEvent = "comment.updated"
	CommentDeletedEvent = "comment.deleted"

	//좋아요 관련
	LikeCreatedEvent = "like.created"
	LikeDeletedEvent = "like.deleted"
)
