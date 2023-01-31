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

// V1TypedLocalObjectReference struct for V1TypedLocalObjectReference
type V1TypedLocalObjectReference struct {
	// APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required. +optional
	ApiGroup *string `json:"apiGroup,omitempty"`
	// Kind is the type of resource being referenced
	Kind *string `json:"kind,omitempty"`
	// Name is the name of resource being referenced
	Name *string `json:"name,omitempty"`
}

// NewV1TypedLocalObjectReference instantiates a new V1TypedLocalObjectReference object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1TypedLocalObjectReference() *V1TypedLocalObjectReference {
	this := V1TypedLocalObjectReference{}
	return &this
}

// NewV1TypedLocalObjectReferenceWithDefaults instantiates a new V1TypedLocalObjectReference object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1TypedLocalObjectReferenceWithDefaults() *V1TypedLocalObjectReference {
	this := V1TypedLocalObjectReference{}
	return &this
}

// GetApiGroup returns the ApiGroup field value if set, zero value otherwise.
func (o *V1TypedLocalObjectReference) GetApiGroup() string {
	if o == nil || o.ApiGroup == nil {
		var ret string
		return ret
	}
	return *o.ApiGroup
}

// GetApiGroupOk returns a tuple with the ApiGroup field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TypedLocalObjectReference) GetApiGroupOk() (*string, bool) {
	if o == nil || o.ApiGroup == nil {
		return nil, false
	}
	return o.ApiGroup, true
}

// HasApiGroup returns a boolean if a field has been set.
func (o *V1TypedLocalObjectReference) HasApiGroup() bool {
	if o != nil && o.ApiGroup != nil {
		return true
	}

	return false
}

// SetApiGroup gets a reference to the given string and assigns it to the ApiGroup field.
func (o *V1TypedLocalObjectReference) SetApiGroup(v string) {
	o.ApiGroup = &v
}

// GetKind returns the Kind field value if set, zero value otherwise.
func (o *V1TypedLocalObjectReference) GetKind() string {
	if o == nil || o.Kind == nil {
		var ret string
		return ret
	}
	return *o.Kind
}

// GetKindOk returns a tuple with the Kind field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TypedLocalObjectReference) GetKindOk() (*string, bool) {
	if o == nil || o.Kind == nil {
		return nil, false
	}
	return o.Kind, true
}

// HasKind returns a boolean if a field has been set.
func (o *V1TypedLocalObjectReference) HasKind() bool {
	if o != nil && o.Kind != nil {
		return true
	}

	return false
}

// SetKind gets a reference to the given string and assigns it to the Kind field.
func (o *V1TypedLocalObjectReference) SetKind(v string) {
	o.Kind = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *V1TypedLocalObjectReference) GetName() string {
	if o == nil || o.Name == nil {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1TypedLocalObjectReference) GetNameOk() (*string, bool) {
	if o == nil || o.Name == nil {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *V1TypedLocalObjectReference) HasName() bool {
	if o != nil && o.Name != nil {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *V1TypedLocalObjectReference) SetName(v string) {
	o.Name = &v
}

func (o V1TypedLocalObjectReference) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.ApiGroup != nil {
		toSerialize["apiGroup"] = o.ApiGroup
	}
	if o.Kind != nil {
		toSerialize["kind"] = o.Kind
	}
	if o.Name != nil {
		toSerialize["name"] = o.Name
	}
	return json.Marshal(toSerialize)
}

type NullableV1TypedLocalObjectReference struct {
	value *V1TypedLocalObjectReference
	isSet bool
}

func (v NullableV1TypedLocalObjectReference) Get() *V1TypedLocalObjectReference {
	return v.value
}

func (v *NullableV1TypedLocalObjectReference) Set(val *V1TypedLocalObjectReference) {
	v.value = val
	v.isSet = true
}

func (v NullableV1TypedLocalObjectReference) IsSet() bool {
	return v.isSet
}

func (v *NullableV1TypedLocalObjectReference) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1TypedLocalObjectReference(val *V1TypedLocalObjectReference) *NullableV1TypedLocalObjectReference {
	return &NullableV1TypedLocalObjectReference{value: val, isSet: true}
}

func (v NullableV1TypedLocalObjectReference) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1TypedLocalObjectReference) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


