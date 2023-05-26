// Code generated by `generate`. DO NOT EDIT.

package kittycad_test

import (
	"fmt"
	"net/url"

	"github.com/kittycad/kittycad.go"
)

// Create a client with your token.
func ExampleNewClient() {
	client, err := kittycad.NewClient("$TOKEN", "your apps user agent")
	if err != nil {
		panic(err)
	}

	// Call the client's methods.
	result, err := client.Meta.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}

// - OR -

// Create a new client with your token parsed from the environment
// variable: `KITTYCAD_API_TOKEN`.
func ExampleNewClientFromEnv() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	// Call the client's methods.
	result, err := client.Meta.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)
}

// Create a client with your token.
func ExampleMetaService_GetSchema() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Meta.GetSchema(); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleMetaService_GetAiPluginManifest() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Meta.GetAiPluginManifest()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleMetaService_Getdata() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Meta.Getdata()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAiService_CreateImageTo3D() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Ai.CreateImageTo3D(kittycad.ImageTypePng, "", []byte("some-binary"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAiService_CreateTextTo3D() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Ai.CreateTextTo3D("", "some-string")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPICallService_GetMetrics() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APICall.GetMetrics("")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPICallService_List() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APICall.List(123, "some-string", "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPICallService_Get() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APICall.Get("some-string")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAppService_GithubCallback() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.App.GithubCallback(""); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleAppService_GithubConsent() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.App.GithubConsent()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAppService_GithubWebhook() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.App.GithubWebhook([]byte("some-binary")); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleAPICallService_ListAsyncOperations() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APICall.ListAsyncOperations(123, "some-string", "", "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPICallService_GetAsyncOperation() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APICall.GetAsyncOperation("some-string")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleHiddenService_AuthEmail() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Hidden.AuthEmail(kittycad.EmailAuthenticationForm{CallbackUrl: kittycad.URL{&url.URL{Scheme: "https", Host: "example.com"}}, Email: "example@example.com"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleHiddenService_AuthEmailCallback() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Hidden.AuthEmailCallback(kittycad.URL{&url.URL{Scheme: "https", Host: "example.com"}}, "example@example.com", "some-string"); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleConstantService_GetPhysics() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Constant.GetPhysics("")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleFileService_CreateCenterOfMass() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.File.CreateCenterOfMass("", []byte("some-binary"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleFileService_CreateConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.File.CreateConversion("", "", []byte("some-binary"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleFileService_CreateDensity() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.File.CreateDensity(123.45, "", []byte("some-binary"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleExecutorService_CreateFileExecution() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Executor.CreateFileExecution(kittycad.CodeLanguageGo, "some-string", []byte("some-binary"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleFileService_CreateMass() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.File.CreateMass(123.45, "", []byte("some-binary"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleFileService_CreateSurfaceArea() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.File.CreateSurfaceArea("", []byte("some-binary"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleFileService_CreateVolume() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.File.CreateVolume("", []byte("some-binary"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleHiddenService_Logout() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Hidden.Logout(); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleModelingService_Cmd() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Modeling.Cmd(kittycad.ModelingCmdReq{Cmd: "", CmdID: kittycad.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), FileID: "some-string"}); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleModelingService_CmdBatch() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Modeling.CmdBatch(kittycad.ModelingCmdReqBatch{Cmds: map[string]kittycad.ModelingCmdReq{"example": {Cmd: "", CmdID: kittycad.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), FileID: "some-string"}}, FileID: "some-string"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleOauth2Service_DeviceAuthRequest() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Oauth2.DeviceAuthRequest(kittycad.DeviceAuthRequestForm{ClientID: kittycad.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8")}); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleOauth2Service_DeviceAuthConfirm() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Oauth2.DeviceAuthConfirm(kittycad.DeviceAuthVerifyParams{UserCode: "some-string"}); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleOauth2Service_DeviceAccessToken() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Oauth2.DeviceAccessToken(kittycad.DeviceAccessTokenRequestForm{ClientID: kittycad.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), DeviceCode: kittycad.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), GrantType: ""}); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleOauth2Service_DeviceAuthVerify() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Oauth2.DeviceAuthVerify("some-string"); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleOauth2Service_ProviderCallback() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Oauth2.ProviderCallback("", "some-string", "some-string"); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleOauth2Service_ProviderConsent() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Oauth2.ProviderConsent("", "some-string")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleMetaService_GetOpenaiSchema() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Meta.GetOpenaiSchema(); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleMetaService_Ping() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Meta.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetAngleConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetAngleConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetAreaConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetAreaConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetCurrentConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetCurrentConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetEnergyConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetEnergyConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetForceConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetForceConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetFrequencyConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetFrequencyConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetLengthConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetLengthConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetMassConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetMassConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetPowerConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetPowerConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetPressureConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetPressureConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetTemperatureConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetTemperatureConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetTorqueConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetTorqueConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUnitService_GetVolumeConversion() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Unit.GetVolumeConversion("", "", 123.45)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_GetSelf() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.GetSelf()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_UpdateSelf() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.UpdateSelf(kittycad.UpdateUser{Company: "some-string", Discord: "some-string", FirstName: "some-string", Github: "some-string", LastName: "some-string", Phone: "+1-555-555-555"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_DeleteSelf() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.User.DeleteSelf(); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleAPICallService_UserList() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APICall.UserList(123, "some-string", "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPICallService_GetForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APICall.GetForUser("some-string")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPITokenService_ListForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APIToken.ListForUser(123, "some-string", "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPITokenService_CreateForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APIToken.CreateForUser()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPITokenService_GetForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APIToken.GetForUser(kittycad.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPITokenService_DeleteForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.APIToken.DeleteForUser(kittycad.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8")); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleUserService_GetSelfExtended() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.GetSelfExtended()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_GetFrontHashSelf() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.GetFrontHashSelf()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_GetOnboardingSelf() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.GetOnboardingSelf()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExamplePaymentService_GetInformationForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Payment.GetInformationForUser()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExamplePaymentService_CreateInformationForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Payment.CreateInformationForUser(kittycad.BillingInfo{Address: kittycad.NewAddress{City: "some-string", Country: "", State: "some-string", Street1: "some-string", Street2: "some-string", UserID: "some-string", Zip: "some-string"}, Name: "some-string", Phone: "+1-555-555-555"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExamplePaymentService_UpdateInformationForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Payment.UpdateInformationForUser(kittycad.BillingInfo{Address: kittycad.NewAddress{City: "some-string", Country: "", State: "some-string", Street1: "some-string", Street2: "some-string", UserID: "some-string", Zip: "some-string"}, Name: "some-string", Phone: "+1-555-555-555"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExamplePaymentService_DeleteInformationForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Payment.DeleteInformationForUser(); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExamplePaymentService_GetBalanceForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Payment.GetBalanceForUser()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExamplePaymentService_CreateIntentForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Payment.CreateIntentForUser()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExamplePaymentService_ListInvoicesForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Payment.ListInvoicesForUser()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExamplePaymentService_ListMethodsForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.Payment.ListMethodsForUser()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExamplePaymentService_DeleteMethodForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Payment.DeleteMethodForUser("some-string"); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExamplePaymentService_ValidateCustomerTaxInformationForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Payment.ValidateCustomerTaxInformationForUser(); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleUserService_GetSessionFor() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.GetSessionFor(kittycad.ParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_List() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.List(123, "some-string", "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_ListExtended() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.ListExtended(123, "some-string", "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_GetExtended() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.GetExtended("some-string")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleUserService_Get() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.User.Get("some-string")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleAPICallService_ListForUser() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	result, err := client.APICall.ListForUser("some-string", 123, "some-string", "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", result)

}

// Create a client with your token.
func ExampleExecutorService_CreateTerm() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Executor.CreateTerm(); err != nil {
		panic(err)
	}

}

// Create a client with your token.
func ExampleModelingService_CommandsWs() {
	client, err := kittycad.NewClientFromEnv("your apps user agent")
	if err != nil {
		panic(err)
	}

	if err := client.Modeling.CommandsWs(); err != nil {
		panic(err)
	}

}
