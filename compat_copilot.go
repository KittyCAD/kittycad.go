package kittycad

// Deprecated: Use MlCopilotMessageUser. This alias retains the original user payload.
type MlCopilotClientMessageCurrentFiles = MlCopilotMessageUser

// Deprecated: Use MlCopilotMessageHeaders.
type MlCopilotClientMessageHeaders = MlCopilotMessageHeaders

// Deprecated: Use MlCopilotMessageListModes.
type MlCopilotClientMessageListModes = MlCopilotMessageListModes

// Deprecated: Use MlCopilotMessageProjectContext.
type MlCopilotClientMessageMlCopilotClientMessageHeaders = MlCopilotMessageProjectContext

// Deprecated: Use MlCopilotMessagePing.
type MlCopilotClientMessagePing = MlCopilotMessagePing

// Deprecated: Use MlCopilotMessageAttachmentResponse. This legacy name represented attachments.
type MlCopilotClientMessageProjectContext = MlCopilotMessageAttachmentResponse

// Deprecated: Use MlCopilotMessageSystem. This alias retains the original system payload.
type MlCopilotClientMessageProjectName = MlCopilotMessageSystem

// Deprecated: Discover available mode strings through list_modes.
type MlCopilotMode = string

const (
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeFast = "fast"
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeThoughtful = "thoughtful"
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeAuto = "auto"
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeZookeeperPro = "zookeeper_pro"
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeZookeeperUltra = "zookeeper_ultra"
)
