package soap

import (
	"encoding/xml"
)

func NewRequestEnvelope() *RequestEnvelope {
	return &RequestEnvelope{
		XMLName: xml.Name{Space: "http://schemas.xmlsoap.org/soap/envelope/", Local: "Envelope"},
	}
}

func (e *RequestEnvelope) SetID(id string) {
	e.Header.ID = id
}

func (e *RequestEnvelope) LoadRPCMethods() {
	e.Body.GetRPCMethodsResponse = &GetRPCMethodsResponse{
		XMLName:    xml.Name{Local: "GetRPCMethodsResponse"},
		MethodList: []string{"GetParameterValues", "SetParameterValues", "Download", "Reboot", "FactoryReset", "AddObject", "DeleteObject", "InformResponse", "RequestXCommand", "TransferCompleteResponse", "GetParameterNames"},
	}
}
