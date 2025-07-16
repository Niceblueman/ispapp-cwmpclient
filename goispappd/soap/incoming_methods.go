package soap

import (
	"encoding/xml"
	"reflect"

	"github.com/sirupsen/logrus"
)

// ACS to CPE soap ResponceEnvelope
func NewResponceEnvelope(logger *logrus.Logger) *ResponceEnvelope {
	return &ResponceEnvelope{}
}

// is valid xml for SOAP requests
func (e *ResponceEnvelope) Load(buf []byte, logger *logrus.Logger) error {
	if err := xml.Unmarshal(buf, e); err != nil {
		logger.Errorf("Failed to unmarshal SOAP response: %v", err)
		return err
	}
	return nil
}

func (e *ResponceEnvelope) GetMethodSwitch() string {
	if e.Body == nil {
		return ""
	}
	if e.Body.GetRPCMethods != nil {
		return "GetRPCMethods"
	}
	if e.Body.GetParameterValues != nil {
		return "GetParameterValues"
	}
	if e.Body.SetParameterValues != nil {
		return "SetParameterValues"
	}
	if e.Body.Download != nil {
		return "Download"
	}
	if e.Body.GetParameterNames != nil {
		return "GetParameterNames"
	}
	if e.Body.Reboot != nil {
		return "Reboot"
	}
	if e.Body.FactoryReset != nil {
		return "FactoryReset"
	}
	if e.Body.AddObject != nil {
		return "AddObject"
	}
	if e.Body.DeleteObject != nil {
		return "DeleteObject"
	}
	if e.Body.InformResponse != nil {
		return "InformResponse"
	}
	if e.Body.TransferCompleteResponse != nil {
		return "TransferCompleteResponse"
	}
	if e.Body.RequestDownloadResponse != nil {
		return "RequestDownloadResponse"
	}
	if e.Body.Fault != nil {
		return "Fault"
	}
	return ""
}
func (e *ResponceEnvelope) GetSize() (int, error) {
	data, err := xml.Marshal(e)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (e *ResponceEnvelope) GetMethod() string {
	v := reflect.ValueOf(e.Body)
	t := v.Type()

	for i := range make([]struct{}, t.NumField()) {
		field := v.Field(i)
		if !field.IsNil() {
			return t.Field(i).Name
		}
	}
	return ""
}

func (e *ResponceEnvelope) GetFault() *FaultResponse {
	if e.Body == nil || e.Body.Fault == nil {
		return nil
	}
	return e.Body.Fault
}
func (e *ResponceEnvelope) GetInformResponse() *InformResponse {
	if e.Body == nil || e.Body.InformResponse == nil {
		return nil
	}
	return e.Body.InformResponse
}
func (e *ResponceEnvelope) GetTransferCompleteResponse() *TransferCompleteResponse {
	if e.Body == nil || e.Body.TransferCompleteResponse == nil {
		return nil
	}
	return e.Body.TransferCompleteResponse
}
func (e *ResponceEnvelope) GetRequestDownloadResponse() *RequestDownloadResponse {
	if e.Body == nil || e.Body.RequestDownloadResponse == nil {
		return nil
	}
	return e.Body.RequestDownloadResponse
}
func (e *ResponceEnvelope) GetGetRPCMethods() *GetRPCMethods {
	if e.Body == nil || e.Body.GetRPCMethods == nil {
		return nil
	}
	return e.Body.GetRPCMethods
}
func (e *ResponceEnvelope) GetGetParameterValues() *GetParameterValues {
	if e.Body == nil || e.Body.GetParameterValues == nil {
		return nil
	}
	return e.Body.GetParameterValues
}
func (e *ResponceEnvelope) GetSetParameterValues() *SetParameterValues {
	if e.Body == nil || e.Body.SetParameterValues == nil {
		return nil
	}
	return e.Body.SetParameterValues
}
func (e *ResponceEnvelope) GetDownload() *Download {
	if e.Body == nil || e.Body.Download == nil {
		return nil
	}
	return e.Body.Download
}
func (e *ResponceEnvelope) GetGetParameterNames() *GetParameterNames {
	if e.Body == nil || e.Body.GetParameterNames == nil {
		return nil
	}
	return e.Body.GetParameterNames
}
func (e *ResponceEnvelope) GetReboot() *Reboot {
	if e.Body == nil || e.Body.Reboot == nil {
		return nil
	}
	return e.Body.Reboot
}
func (e *ResponceEnvelope) GetFactoryReset() *FactoryReset {
	if e.Body == nil || e.Body.FactoryReset == nil {
		return nil
	}
	return e.Body.FactoryReset
}
func (e *ResponceEnvelope) GetAddObject() *AddObject {
	if e.Body == nil || e.Body.AddObject == nil {
		return nil
	}
	return e.Body.AddObject
}
func (e *ResponceEnvelope) GetDeleteObject() *DeleteObject {
	if e.Body == nil || e.Body.DeleteObject == nil {
		return nil
	}
	return e.Body.DeleteObject
}
