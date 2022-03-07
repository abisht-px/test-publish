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

// ControllersCreateAccountRequest struct for ControllersCreateAccountRequest
type ControllersCreateAccountRequest struct {
	Name *string `json:"name,omitempty"`
}

// NewControllersCreateAccountRequest instantiates a new ControllersCreateAccountRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewControllersCreateAccountRequest() *ControllersCreateAccountRequest {
	this := ControllersCreateAccountRequest{}
	return &this
}

// NewControllersCreateAccountRequestWithDefaults instantiates a new ControllersCreateAccountRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewControllersCreateAccountRequestWithDefaults() *ControllersCreateAccountRequest {
	this := ControllersCreateAccountRequest{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ControllersCreateAccountRequest) GetName() string {
	if o == nil || o.Name == nil {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ControllersCreateAccountRequest) GetNameOk() (*string, bool) {
	if o == nil || o.Name == nil {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ControllersCreateAccountRequest) HasName() bool {
	if o != nil && o.Name != nil {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ControllersCreateAccountRequest) SetName(v string) {
	o.Name = &v
}

func (o ControllersCreateAccountRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Name != nil {
		toSerialize["name"] = o.Name
	}
	return json.Marshal(toSerialize)
}

type NullableControllersCreateAccountRequest struct {
	value *ControllersCreateAccountRequest
	isSet bool
}

func (v NullableControllersCreateAccountRequest) Get() *ControllersCreateAccountRequest {
	return v.value
}

func (v *NullableControllersCreateAccountRequest) Set(val *ControllersCreateAccountRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableControllersCreateAccountRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableControllersCreateAccountRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableControllersCreateAccountRequest(val *ControllersCreateAccountRequest) *NullableControllersCreateAccountRequest {
	return &NullableControllersCreateAccountRequest{value: val, isSet: true}
}

func (v NullableControllersCreateAccountRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableControllersCreateAccountRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

