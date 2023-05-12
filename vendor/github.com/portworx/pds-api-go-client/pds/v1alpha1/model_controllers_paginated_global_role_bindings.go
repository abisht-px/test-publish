/*
PDS API

Portworx Data Services API Server

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pds

import (
	"encoding/json"
)

// ControllersPaginatedGlobalRoleBindings struct for ControllersPaginatedGlobalRoleBindings
type ControllersPaginatedGlobalRoleBindings struct {
	Data []ModelsLegacyGlobalBinding `json:"data,omitempty"`
}

// NewControllersPaginatedGlobalRoleBindings instantiates a new ControllersPaginatedGlobalRoleBindings object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewControllersPaginatedGlobalRoleBindings() *ControllersPaginatedGlobalRoleBindings {
	this := ControllersPaginatedGlobalRoleBindings{}
	return &this
}

// NewControllersPaginatedGlobalRoleBindingsWithDefaults instantiates a new ControllersPaginatedGlobalRoleBindings object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewControllersPaginatedGlobalRoleBindingsWithDefaults() *ControllersPaginatedGlobalRoleBindings {
	this := ControllersPaginatedGlobalRoleBindings{}
	return &this
}

// GetData returns the Data field value if set, zero value otherwise.
func (o *ControllersPaginatedGlobalRoleBindings) GetData() []ModelsLegacyGlobalBinding {
	if o == nil || o.Data == nil {
		var ret []ModelsLegacyGlobalBinding
		return ret
	}
	return o.Data
}

// GetDataOk returns a tuple with the Data field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ControllersPaginatedGlobalRoleBindings) GetDataOk() ([]ModelsLegacyGlobalBinding, bool) {
	if o == nil || o.Data == nil {
		return nil, false
	}
	return o.Data, true
}

// HasData returns a boolean if a field has been set.
func (o *ControllersPaginatedGlobalRoleBindings) HasData() bool {
	if o != nil && o.Data != nil {
		return true
	}

	return false
}

// SetData gets a reference to the given []ModelsLegacyGlobalBinding and assigns it to the Data field.
func (o *ControllersPaginatedGlobalRoleBindings) SetData(v []ModelsLegacyGlobalBinding) {
	o.Data = v
}

func (o ControllersPaginatedGlobalRoleBindings) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Data != nil {
		toSerialize["data"] = o.Data
	}
	return json.Marshal(toSerialize)
}

type NullableControllersPaginatedGlobalRoleBindings struct {
	value *ControllersPaginatedGlobalRoleBindings
	isSet bool
}

func (v NullableControllersPaginatedGlobalRoleBindings) Get() *ControllersPaginatedGlobalRoleBindings {
	return v.value
}

func (v *NullableControllersPaginatedGlobalRoleBindings) Set(val *ControllersPaginatedGlobalRoleBindings) {
	v.value = val
	v.isSet = true
}

func (v NullableControllersPaginatedGlobalRoleBindings) IsSet() bool {
	return v.isSet
}

func (v *NullableControllersPaginatedGlobalRoleBindings) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableControllersPaginatedGlobalRoleBindings(val *ControllersPaginatedGlobalRoleBindings) *NullableControllersPaginatedGlobalRoleBindings {
	return &NullableControllersPaginatedGlobalRoleBindings{value: val, isSet: true}
}

func (v NullableControllersPaginatedGlobalRoleBindings) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableControllersPaginatedGlobalRoleBindings) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


