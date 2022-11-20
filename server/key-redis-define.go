package server

// Session Cache
// session:useruuid:token
const KeySessionFmt = "session:%s:%s"

// session:useruuid:token
const KeySessionCacheRefreshFmt = "session-refresh:%s:%s"

// Session Register
// session-reg:session-uuid
const KeyHsetSessionRegFmt = "session-reg"
