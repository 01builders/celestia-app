package types

// NewUpdateMinfeeParamsEvent returns a new EventUpdateMinfeeParams
func NewUpdateMinfeeParamsEvent(authority string, params Params) *EventUpdateMinfeeParams {
	return &EventUpdateMinfeeParams{
		Signer: authority,
		Params: params,
	}
}

// NewEventUpdateMinfeeParams creates a new instance of EventUpdateMinfeeParams
func NewEventUpdateMinfeeParams(signer string, params Params) *EventUpdateMinfeeParams {
	return &EventUpdateMinfeeParams{
		Signer: signer,
		Params: params,
	}
}
