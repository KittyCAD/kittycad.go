// Code generated by `generate`. DO NOT EDIT.

package kittycad

import "net/http"

// Client which conforms to the OpenAPI v3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.example.com for example.
	server string

	// Client is the *http.Client for performing requests.
	client *http.Client

	// token is the API token used for authentication.
	token string

	// Ai: AI uses machine learning to generate 3D meshes.
	Ai *AiService
	// APICall: API calls that have been performed by users can be queried by the API. This is helpful for debugging as well as billing.
	APICall *APICallService
	// APIToken: API tokens allow users to call the API outside of their session token that is used as a cookie in the user interface. Users can create, delete, and list their API tokens. But, of course, you need an API token to do this, so first be sure to generate one in the account UI.
	APIToken *APITokenService
	// App: Endpoints for third party app grant flows.
	App *AppService
	// Beta: Beta API endpoints. We will not charge for these endpoints while they are in beta.
	Beta *BetaService
	// Constant: Constants. These are helpful as helpers.
	Constant *ConstantService
	// Drawing: Drawing API for updating your 3D files using the KittyCAD engine.
	Drawing *DrawingService
	// File: CAD file operations. Create, get, and list CAD file conversions. More endpoints will be added here in the future as we build out transforms, etc on CAD models.
	File *FileService
	// Hidden: Hidden API endpoints that should not show up in the docs.
	Hidden *HiddenService
	// Meta: Meta information about the API.
	Meta *MetaService
	// Oauth2: Endpoints that implement OAuth 2.0 grant flows.
	Oauth2 *Oauth2Service
	// Payment: Operations around payments and billing.
	Payment *PaymentService
	// Session: Sessions allow users to call the API from their session cookie in the browser.
	Session *SessionService
	// Unit: Unit conversion operations.
	Unit *UnitService
	// User: A user is someone who uses the KittyCAD API. Here, we can create, delete, and list users. We can also get information about a user. Operations will only be authorized if the user is requesting information about themselves.
	User *UserService
}

// AiService: AI uses machine learning to generate 3D meshes.
type AiService service

// APICallService: API calls that have been performed by users can be queried by the API. This is helpful for debugging as well as billing.
type APICallService service

// APITokenService: API tokens allow users to call the API outside of their session token that is used as a cookie in the user interface. Users can create, delete, and list their API tokens. But, of course, you need an API token to do this, so first be sure to generate one in the account UI.
type APITokenService service

// AppService: Endpoints for third party app grant flows.
type AppService service

// BetaService: Beta API endpoints. We will not charge for these endpoints while they are in beta.
type BetaService service

// ConstantService: Constants. These are helpful as helpers.
type ConstantService service

// DrawingService: Drawing API for updating your 3D files using the KittyCAD engine.
type DrawingService service

// FileService: CAD file operations. Create, get, and list CAD file conversions. More endpoints will be added here in the future as we build out transforms, etc on CAD models.
type FileService service

// HiddenService: Hidden API endpoints that should not show up in the docs.
type HiddenService service

// MetaService: Meta information about the API.
type MetaService service

// Oauth2Service: Endpoints that implement OAuth 2.0 grant flows.
type Oauth2Service service

// PaymentService: Operations around payments and billing.
type PaymentService service

// SessionService: Sessions allow users to call the API from their session cookie in the browser.
type SessionService service

// UnitService: Unit conversion operations.
type UnitService service

// UserService: A user is someone who uses the KittyCAD API. Here, we can create, delete, and list users. We can also get information about a user. Operations will only be authorized if the user is requesting information about themselves.
type UserService service
