package ws

const (
	onlineExpireSeconds = 24 * 3600
	messageStreamMaxLen = 1000

	presenceKeyPrefix     = "presence:"
	roomPresenceKeyPrefix = "presence:room:"
	messageStreamPrefix   = "stream:message:"
	aiStreamPrefix        = "stream:ai:"
	unreadKeyPrefix       = "unread:"
	aiChatKeyPrefix       = "aichat:"
)
