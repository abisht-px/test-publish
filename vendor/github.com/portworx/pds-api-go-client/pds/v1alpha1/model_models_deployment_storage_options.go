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

// ModelsDeploymentStorageOptions struct for ModelsDeploymentStorageOptions
type ModelsDeploymentStorageOptions struct {
	Fg *bool `json:"fg,omitempty"`
	Fs *string `json:"fs,omitempty"`
	Repl *int32 `json:"repl,omitempty"`
	Secure *bool `json:"secure,omitempty"`
}

// NewModelsDeploymentStorageOptions instantiates a new ModelsDeploymentStorageOptions object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewModelsDeploymentStorageOptions() *ModelsDeploymentStorageOptions {
	this := ModelsDeploymentStorageOptions{}
	return &this
}

// NewModelsDeploymentStorageOptionsWithDefaults instantiates a new ModelsDeploymentStorageOptions object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewModelsDeploymentStorageOptionsWithDefaults() *ModelsDeploymentStorageOptions {
	this := ModelsDeploymentStorageOptions{}
	return &this
}

// GetFg returns the Fg field value if set, zero value otherwise.
func (o *ModelsDeploymentStorageOptions) GetFg() bool {
	if o == nil || o.Fg == nil {
		var ret bool
		return ret
	}
	return *o.Fg
}

// GetFgOk returns a tuple with the Fg field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ModelsDeploymentStorageOptions) GetFgOk() (*bool, bool) {
	if o == nil || o.Fg == nil {
		return nil, false
	}
	return o.Fg, true
}

// HasFg returns a boolean if a field has been set.
func (o *ModelsDeploymentStorageOptions) HasFg() bool {
	if o != nil && o.Fg != nil {
		return true
	}

	return false
}

// SetFg gets a reference to the given bool and assigns it to the Fg field.
func (o *ModelsDeploymentStorageOptions) SetFg(v bool) {
	o.Fg = &v
}

// GetFs returns the Fs field value if set, zero value otherwise.
func (o *ModelsDeploymentStorageOptions) GetFs() string {
	if o == nil || o.Fs == nil {
		var ret string
		return ret
	}
	return *o.Fs
}

// GetFsOk returns a tuple with the Fs field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ModelsDeploymentStorageOptions) GetFsOk() (*string, bool) {
	if o == nil || o.Fs == nil {
		return nil, false
	}
	return o.Fs, true
}

// HasFs returns a boolean if a field has been set.
func (o *ModelsDeploymentStorageOptions) HasFs() bool {
	if o != nil && o.Fs != nil {
		return true
	}

	return false
}

// SetFs gets a reference to the given string and assigns it to the Fs field.
func (o *ModelsDeploymentStorageOptions) SetFs(v string) {
	o.Fs = &v
}

// GetRepl returns the Repl field value if set, zero value otherwise.
func (o *ModelsDeploymentStorageOptions) GetRepl() int32 {
	if o == nil || o.Repl == nil {
		var ret int32
		return ret
	}
	return *o.Repl
}

// GetReplOk returns a tuple with the Repl field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ModelsDeploymentStorageOptions) GetReplOk() (*int32, bool) {
	if o == nil || o.Repl == nil {
		return nil, false
	}
	return o.Repl, true
}

// HasRepl returns a boolean if a field has been set.
func (o *ModelsDeploymentStorageOptions) HasRepl() bool {
	if o != nil && o.Repl != nil {
		return true
	}

	return false
}

// SetRepl gets a reference to the given int32 and assigns it to the Repl field.
func (o *ModelsDeploymentStorageOptions) SetRepl(v int32) {
	o.Repl = &v
}

// GetSecure returns the Secure field value if set, zero value otherwise.
func (o *ModelsDeploymentStorageOptions) GetSecure() bool {
	if o == nil || o.Secure == nil {
		var ret bool
		return ret
	}
	return *o.Secure
}

// GetSecureOk returns a tuple with the Secure field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ModelsDeploymentStorageOptions) GetSecureOk() (*bool, bool) {
	if o == nil || o.Secure == nil {
		return nil, false
	}
	return o.Secure, true
}

// HasSecure returns a boolean if a field has been set.
func (o *ModelsDeploymentStorageOptions) HasSecure() bool {
	if o != nil && o.Secure != nil {
		return true
	}

	return false
}

// SetSecure gets a reference to the given bool and assigns it to the Secure field.
func (o *ModelsDeploymentStorageOptions) SetSecure(v bool) {
	o.Secure = &v
}

func (o ModelsDeploymentStorageOptions) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Fg != nil {
		toSerialize["fg"] = o.Fg
	}
	if o.Fs != nil {
		toSerialize["fs"] = o.Fs
	}
	if o.Repl != nil {
		toSerialize["repl"] = o.Repl
	}
	if o.Secure != nil {
		toSerialize["secure"] = o.Secure
	}
	return json.Marshal(toSerialize)
}

type NullableModelsDeploymentStorageOptions struct {
	value *ModelsDeploymentStorageOptions
	isSet bool
}

func (v NullableModelsDeploymentStorageOptions) Get() *ModelsDeploymentStorageOptions {
	return v.value
}

func (v *NullableModelsDeploymentStorageOptions) Set(val *ModelsDeploymentStorageOptions) {
	v.value = val
	v.isSet = true
}

func (v NullableModelsDeploymentStorageOptions) IsSet() bool {
	return v.isSet
}

func (v *NullableModelsDeploymentStorageOptions) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableModelsDeploymentStorageOptions(val *ModelsDeploymentStorageOptions) *NullableModelsDeploymentStorageOptions {
	return &NullableModelsDeploymentStorageOptions{value: val, isSet: true}
}

func (v NullableModelsDeploymentStorageOptions) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableModelsDeploymentStorageOptions) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


