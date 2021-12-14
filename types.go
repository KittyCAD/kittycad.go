// Code generated by `generate`. DO NOT EDIT.

package kittycad

import "time"

// ValidFileType is the type definition for a ValidFileType.
type ValidFileType string

const (
	// VALID_FILE_TYPE_STEP represents the ValidFileType `"step"`.
	VALID_FILE_TYPE_STEP ValidFileType = "step"
	// VALID_FILE_TYPE_OBJ represents the ValidFileType `"obj"`.
	VALID_FILE_TYPE_OBJ ValidFileType = "obj"
	// VALID_FILE_TYPE_STL represents the ValidFileType `"stl"`.
	VALID_FILE_TYPE_STL ValidFileType = "stl"
	// VALID_FILE_TYPE_DXF represents the ValidFileType `"dxf"`.
	VALID_FILE_TYPE_DXF ValidFileType = "dxf"
	// VALID_FILE_TYPE_DWG represents the ValidFileType `"dwg"`.
	VALID_FILE_TYPE_DWG ValidFileType = "dwg"
)

// ValidFileTypes is the collection of all ValidFileType values.
var ValidFileTypes = []ValidFileType{
	VALID_FILE_TYPE_STEP,
	VALID_FILE_TYPE_OBJ,
	VALID_FILE_TYPE_STL,
	VALID_FILE_TYPE_DXF,
	VALID_FILE_TYPE_DWG,
}

// AuthSession is the type definition for a AuthSession.
type AuthSession struct {
	// UserID is the user's id.
	UserID string `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	// CreatedAt is the date and time the session/request was created.
	CreatedAt time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	// Email is the user's email address.
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
	// ID is the id of the session.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// IPAddress is the IP address the request originated from.
	IPAddress string `json:"ip_address,omitempty" yaml:"ip_address,omitempty"`
	// IsValid is if the token is valid.
	IsValid bool `json:"is_valid,omitempty" yaml:"is_valid,omitempty"`
	// Token is the user's token.
	Token string `json:"token,omitempty" yaml:"token,omitempty"`
}

// Environment is the type of environment.
type Environment string

const (
	// ENVIRONMENT_DEVELOPMENT represents the Environment `"DEVELOPMENT"`.
	ENVIRONMENT_DEVELOPMENT Environment = "DEVELOPMENT"
	// ENVIRONMENT_PREVIEW represents the Environment `"PREVIEW"`.
	ENVIRONMENT_PREVIEW Environment = "PREVIEW"
	// ENVIRONMENT_PRODUCTION represents the Environment `"PRODUCTION"`.
	ENVIRONMENT_PRODUCTION Environment = "PRODUCTION"
)

// Environments is the collection of all Environment values.
var Environments = []Environment{
	ENVIRONMENT_DEVELOPMENT,
	ENVIRONMENT_PREVIEW,
	ENVIRONMENT_PRODUCTION,
}

// ErrorMessage is the type definition for a ErrorMessage.
type ErrorMessage struct {
	// Message is the message.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// FileConversion is the type definition for a FileConversion.
type FileConversion struct {
	SrcFormat ValidFileType `json:"src_format,omitempty" yaml:"src_format,omitempty"`
	// Status is the status of the file conversion.
	Status FileConversionStatus `json:"status,omitempty" yaml:"status,omitempty"`
	// CompletedAt is the date and time the file conversion was completed.
	CompletedAt time.Time `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	// CreatedAt is the date and time the file conversion was created.
	CreatedAt time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	// ID is the id of the file conversion.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// Output is the converted file, base64 encoded.
	Output       string        `json:"output,omitempty" yaml:"output,omitempty"`
	OutputFormat ValidFileType `json:"output_format,omitempty" yaml:"output_format,omitempty"`
}

// FileConversionStatus is the status of the file conversion.
type FileConversionStatus string

const (
	// FILE_CONVERSION_STATUS_QUEUED represents the FileConversionStatus `"Queued"`.
	FILE_CONVERSION_STATUS_QUEUED FileConversionStatus = "Queued"
	// FILE_CONVERSION_STATUS_UPLOADED represents the FileConversionStatus `"Uploaded"`.
	FILE_CONVERSION_STATUS_UPLOADED FileConversionStatus = "Uploaded"
	// FILE_CONVERSION_STATUS_IN_PROGRESS represents the FileConversionStatus `"In Progress"`.
	FILE_CONVERSION_STATUS_IN_PROGRESS FileConversionStatus = "In Progress"
	// FILE_CONVERSION_STATUS_COMPLETED represents the FileConversionStatus `"Completed"`.
	FILE_CONVERSION_STATUS_COMPLETED FileConversionStatus = "Completed"
	// FILE_CONVERSION_STATUS_FAILED represents the FileConversionStatus `"Failed"`.
	FILE_CONVERSION_STATUS_FAILED FileConversionStatus = "Failed"
)

// FileConversionStatuses is the collection of all FileConversionStatus values.
var FileConversionStatuses = []FileConversionStatus{
	FILE_CONVERSION_STATUS_QUEUED,
	FILE_CONVERSION_STATUS_UPLOADED,
	FILE_CONVERSION_STATUS_IN_PROGRESS,
	FILE_CONVERSION_STATUS_COMPLETED,
	FILE_CONVERSION_STATUS_FAILED,
}

// InstanceMetadata is the type definition for a InstanceMetadata.
type InstanceMetadata struct {
	// Image is the image that was used as the base of the instance.
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	// IPAddress is the IP address of the instance.
	IPAddress string `json:"ip_address,omitempty" yaml:"ip_address,omitempty"`
	// MachineType is the machine type of the instance.
	MachineType string `json:"machine_type,omitempty" yaml:"machine_type,omitempty"`
	// Name is the name of the instance.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// GitHash is the git hash of the code the server was built from.
	GitHash string `json:"git_hash,omitempty" yaml:"git_hash,omitempty"`
	// Hostname is the hostname of the instance.
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// ID is the id of the instance.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// Zone is the zone of the instance.
	Zone string `json:"zone,omitempty" yaml:"zone,omitempty"`
	// CPUPlatform is the CPU platform of the instance.
	CPUPlatform string `json:"cpu_platform,omitempty" yaml:"cpu_platform,omitempty"`
	// Description is the description of the instance.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Environment is the type of environment.
	Environment Environment `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// Message is the type definition for a Message.
type Message struct {
	// Message is the message.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}
