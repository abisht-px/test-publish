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

// ControllersDeploymentTargetHeartbeatRequest struct for ControllersDeploymentTargetHeartbeatRequest
type ControllersDeploymentTargetHeartbeatRequest struct {
	ClusterId *string `json:"cluster_id,omitempty"`
}

// NewControllersDeploymentTargetHeartbeatRequest instantiates a new ControllersDeploymentTargetHeartbeatRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewControllersDeploymentTargetHeartbeatRequest() *ControllersDeploymentTargetHeartbeatRequest {
	this := ControllersDeploymentTargetHeartbeatRequest{}
	return &this
}

// NewControllersDeploymentTargetHeartbeatRequestWithDefaults instantiates a new ControllersDeploymentTargetHeartbeatRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewControllersDeploymentTargetHeartbeatRequestWithDefaults() *ControllersDeploymentTargetHeartbeatRequest {
	this := ControllersDeploymentTargetHeartbeatRequest{}
	return &this
}

// GetClusterId returns the ClusterId field value if set, zero value otherwise.
func (o *ControllersDeploymentTargetHeartbeatRequest) GetClusterId() string {
	if o == nil || o.ClusterId == nil {
		var ret string
		return ret
	}
	return *o.ClusterId
}

// GetClusterIdOk returns a tuple with the ClusterId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ControllersDeploymentTargetHeartbeatRequest) GetClusterIdOk() (*string, bool) {
	if o == nil || o.ClusterId == nil {
		return nil, false
	}
	return o.ClusterId, true
}

// HasClusterId returns a boolean if a field has been set.
func (o *ControllersDeploymentTargetHeartbeatRequest) HasClusterId() bool {
	if o != nil && o.ClusterId != nil {
		return true
	}

	return false
}

// SetClusterId gets a reference to the given string and assigns it to the ClusterId field.
func (o *ControllersDeploymentTargetHeartbeatRequest) SetClusterId(v string) {
	o.ClusterId = &v
}

func (o ControllersDeploymentTargetHeartbeatRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.ClusterId != nil {
		toSerialize["cluster_id"] = o.ClusterId
	}
	return json.Marshal(toSerialize)
}

type NullableControllersDeploymentTargetHeartbeatRequest struct {
	value *ControllersDeploymentTargetHeartbeatRequest
	isSet bool
}

func (v NullableControllersDeploymentTargetHeartbeatRequest) Get() *ControllersDeploymentTargetHeartbeatRequest {
	return v.value
}

func (v *NullableControllersDeploymentTargetHeartbeatRequest) Set(val *ControllersDeploymentTargetHeartbeatRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableControllersDeploymentTargetHeartbeatRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableControllersDeploymentTargetHeartbeatRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableControllersDeploymentTargetHeartbeatRequest(val *ControllersDeploymentTargetHeartbeatRequest) *NullableControllersDeploymentTargetHeartbeatRequest {
	return &NullableControllersDeploymentTargetHeartbeatRequest{value: val, isSet: true}
}

func (v NullableControllersDeploymentTargetHeartbeatRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableControllersDeploymentTargetHeartbeatRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

