// Code generated by Speakeasy (https://speakeasyapi.com). DO NOT EDIT.

package components

type V2Account struct {
	Address          string              `json:"address"`
	Metadata         map[string]string   `json:"metadata"`
	Volumes          map[string]V2Volume `json:"volumes,omitempty"`
	EffectiveVolumes map[string]V2Volume `json:"effectiveVolumes,omitempty"`
}

func (o *V2Account) GetAddress() string {
	if o == nil {
		return ""
	}
	return o.Address
}

func (o *V2Account) GetMetadata() map[string]string {
	if o == nil {
		return map[string]string{}
	}
	return o.Metadata
}

func (o *V2Account) GetVolumes() map[string]V2Volume {
	if o == nil {
		return nil
	}
	return o.Volumes
}

func (o *V2Account) GetEffectiveVolumes() map[string]V2Volume {
	if o == nil {
		return nil
	}
	return o.EffectiveVolumes
}
