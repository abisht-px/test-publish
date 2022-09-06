/*
PDS API

Portworx Data Services API Server

API version: 3.0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pds

import (
	"encoding/json"
)

// ControllersTargetClusterConfigResponse struct for ControllersTargetClusterConfigResponse
type ControllersTargetClusterConfigResponse struct {
	ObservabilityUrl *string `json:"observability_url,omitempty"`
	TeleportCaPin *string `json:"teleport_ca_pin,omitempty"`
	TeleportJoinToken *string `json:"teleport_join_token,omitempty"`
	TeleportProxyAddr *string `json:"teleport_proxy_addr,omitempty"`
}

// NewControllersTargetClusterConfigResponse instantiates a new ControllersTargetClusterConfigResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewControllersTargetClusterConfigResponse() *ControllersTargetClusterConfigResponse {
	this := ControllersTargetClusterConfigResponse{}
	return &this
}

// NewControllersTargetClusterConfigResponseWithDefaults instantiates a new ControllersTargetClusterConfigResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewControllersTargetClusterConfigResponseWithDefaults() *ControllersTargetClusterConfigResponse {
	this := ControllersTargetClusterConfigResponse{}
	return &this
}

// GetObservabilityUrl returns the ObservabilityUrl field value if set, zero value otherwise.
func (o *ControllersTargetClusterConfigResponse) GetObservabilityUrl() string {
	if o == nil || o.ObservabilityUrl == nil {
		var ret string
		return ret
	}
	return *o.ObservabilityUrl
}

// GetObservabilityUrlOk returns a tuple with the ObservabilityUrl field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ControllersTargetClusterConfigResponse) GetObservabilityUrlOk() (*string, bool) {
	if o == nil || o.ObservabilityUrl == nil {
		return nil, false
	}
	return o.ObservabilityUrl, true
}

// HasObservabilityUrl returns a boolean if a field has been set.
func (o *ControllersTargetClusterConfigResponse) HasObservabilityUrl() bool {
	if o != nil && o.ObservabilityUrl != nil {
		return true
	}

	return false
}

// SetObservabilityUrl gets a reference to the given string and assigns it to the ObservabilityUrl field.
func (o *ControllersTargetClusterConfigResponse) SetObservabilityUrl(v string) {
	o.ObservabilityUrl = &v
}

// GetTeleportCaPin returns the TeleportCaPin field value if set, zero value otherwise.
func (o *ControllersTargetClusterConfigResponse) GetTeleportCaPin() string {
	if o == nil || o.TeleportCaPin == nil {
		var ret string
		return ret
	}
	return *o.TeleportCaPin
}

// GetTeleportCaPinOk returns a tuple with the TeleportCaPin field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ControllersTargetClusterConfigResponse) GetTeleportCaPinOk() (*string, bool) {
	if o == nil || o.TeleportCaPin == nil {
		return nil, false
	}
	return o.TeleportCaPin, true
}

// HasTeleportCaPin returns a boolean if a field has been set.
func (o *ControllersTargetClusterConfigResponse) HasTeleportCaPin() bool {
	if o != nil && o.TeleportCaPin != nil {
		return true
	}

	return false
}

// SetTeleportCaPin gets a reference to the given string and assigns it to the TeleportCaPin field.
func (o *ControllersTargetClusterConfigResponse) SetTeleportCaPin(v string) {
	o.TeleportCaPin = &v
}

// GetTeleportJoinToken returns the TeleportJoinToken field value if set, zero value otherwise.
func (o *ControllersTargetClusterConfigResponse) GetTeleportJoinToken() string {
	if o == nil || o.TeleportJoinToken == nil {
		var ret string
		return ret
	}
	return *o.TeleportJoinToken
}

// GetTeleportJoinTokenOk returns a tuple with the TeleportJoinToken field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ControllersTargetClusterConfigResponse) GetTeleportJoinTokenOk() (*string, bool) {
	if o == nil || o.TeleportJoinToken == nil {
		return nil, false
	}
	return o.TeleportJoinToken, true
}

// HasTeleportJoinToken returns a boolean if a field has been set.
func (o *ControllersTargetClusterConfigResponse) HasTeleportJoinToken() bool {
	if o != nil && o.TeleportJoinToken != nil {
		return true
	}

	return false
}

// SetTeleportJoinToken gets a reference to the given string and assigns it to the TeleportJoinToken field.
func (o *ControllersTargetClusterConfigResponse) SetTeleportJoinToken(v string) {
	o.TeleportJoinToken = &v
}

// GetTeleportProxyAddr returns the TeleportProxyAddr field value if set, zero value otherwise.
func (o *ControllersTargetClusterConfigResponse) GetTeleportProxyAddr() string {
	if o == nil || o.TeleportProxyAddr == nil {
		var ret string
		return ret
	}
	return *o.TeleportProxyAddr
}

// GetTeleportProxyAddrOk returns a tuple with the TeleportProxyAddr field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ControllersTargetClusterConfigResponse) GetTeleportProxyAddrOk() (*string, bool) {
	if o == nil || o.TeleportProxyAddr == nil {
		return nil, false
	}
	return o.TeleportProxyAddr, true
}

// HasTeleportProxyAddr returns a boolean if a field has been set.
func (o *ControllersTargetClusterConfigResponse) HasTeleportProxyAddr() bool {
	if o != nil && o.TeleportProxyAddr != nil {
		return true
	}

	return false
}

// SetTeleportProxyAddr gets a reference to the given string and assigns it to the TeleportProxyAddr field.
func (o *ControllersTargetClusterConfigResponse) SetTeleportProxyAddr(v string) {
	o.TeleportProxyAddr = &v
}

func (o ControllersTargetClusterConfigResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.ObservabilityUrl != nil {
		toSerialize["observability_url"] = o.ObservabilityUrl
	}
	if o.TeleportCaPin != nil {
		toSerialize["teleport_ca_pin"] = o.TeleportCaPin
	}
	if o.TeleportJoinToken != nil {
		toSerialize["teleport_join_token"] = o.TeleportJoinToken
	}
	if o.TeleportProxyAddr != nil {
		toSerialize["teleport_proxy_addr"] = o.TeleportProxyAddr
	}
	return json.Marshal(toSerialize)
}

type NullableControllersTargetClusterConfigResponse struct {
	value *ControllersTargetClusterConfigResponse
	isSet bool
}

func (v NullableControllersTargetClusterConfigResponse) Get() *ControllersTargetClusterConfigResponse {
	return v.value
}

func (v *NullableControllersTargetClusterConfigResponse) Set(val *ControllersTargetClusterConfigResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableControllersTargetClusterConfigResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableControllersTargetClusterConfigResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableControllersTargetClusterConfigResponse(val *ControllersTargetClusterConfigResponse) *NullableControllersTargetClusterConfigResponse {
	return &NullableControllersTargetClusterConfigResponse{value: val, isSet: true}
}

func (v NullableControllersTargetClusterConfigResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableControllersTargetClusterConfigResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

