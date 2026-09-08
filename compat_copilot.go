package kittycad

// MlCopilotClientMessageCurrentFiles preserves the legacy SDK name.
//
// Deprecated: Use MlCopilotMessageUser. This alias retains the original user payload.
type MlCopilotClientMessageCurrentFiles = MlCopilotMessageUser

// MlCopilotClientMessageHeaders preserves the legacy SDK name.
//
// Deprecated: Use MlCopilotMessageHeaders.
type MlCopilotClientMessageHeaders = MlCopilotMessageHeaders

// MlCopilotClientMessageListModes preserves the legacy SDK name.
//
// Deprecated: Use MlCopilotMessageListModes.
type MlCopilotClientMessageListModes = MlCopilotMessageListModes

// MlCopilotClientMessageMlCopilotClientMessageHeaders preserves the legacy SDK name.
//
// Deprecated: Use MlCopilotMessageProjectContext.
type MlCopilotClientMessageMlCopilotClientMessageHeaders = MlCopilotMessageProjectContext

// MlCopilotClientMessagePing preserves the legacy SDK name.
//
// Deprecated: Use MlCopilotMessagePing.
type MlCopilotClientMessagePing = MlCopilotMessagePing

// MlCopilotClientMessageProjectContext preserves the legacy SDK name.
//
// Deprecated: Use MlCopilotMessageAttachmentResponse. This legacy name represented attachments.
type MlCopilotClientMessageProjectContext = MlCopilotMessageAttachmentResponse

// MlCopilotClientMessageProjectName preserves the legacy SDK name.
//
// Deprecated: Use MlCopilotMessageSystem. This alias retains the original system payload.
type MlCopilotClientMessageProjectName = MlCopilotMessageSystem

// MlCopilotMode preserves the legacy SDK name.
//
// Deprecated: Discover available mode strings through list_modes.
type MlCopilotMode = string

const (
	// MlCopilotModeFast preserves the legacy SDK name.
	//
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeFast = "fast"
	// MlCopilotModeThoughtful preserves the legacy SDK name.
	//
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeThoughtful = "thoughtful"
	// MlCopilotModeAuto preserves the legacy SDK name.
	//
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeAuto = "auto"
	// MlCopilotModeZookeeperPro preserves the legacy SDK name.
	//
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeZookeeperPro = "zookeeper_pro"
	// MlCopilotModeZookeeperUltra preserves the legacy SDK name.
	//
	// Deprecated: Discover available mode strings through list_modes.
	MlCopilotModeZookeeperUltra = "zookeeper_ultra"
)
