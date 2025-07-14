package soap

import (
	"encoding/xml"
)

// RequestEnvelope represents sent messages from APE to ACS
type RequestEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Header  struct {
		ID string `xml:"ID"`
	} `xml:"Header"`
	Body struct {
		TransferComplete           *TransferComplete           `xml:"TransferComplete,omitempty"`
		RequestDownload            *RequestDownload            `xml:"RequestDownload,omitempty"`
		AutonomousTransferComplete *AutonomousTransferComplete `xml:"AutonomousTransferComplete,omitempty"`
		ScheduleInform             *ScheduleInform             `xml:"ScheduleInform,omitempty"`
		SetVouchers                *SetVouchers                `xml:"SetVouchers,omitempty"`
		GetOptions                 *GetOptions                 `xml:"GetOptions,omitempty"`
		Inform                     *Inform                     `xml:"Inform,omitempty"`
		Fault                      *Fault                      `xml:"Fault,omitempty"`
		DownloadResponse           *DownloadResponse           `xml:"DownloadResponse,omitempty"`
		GetRPCMethodsResponse      *GetRPCMethodsResponse      `xml:"GetRPCMethodsResponse,omitempty"`
		SetParameterValuesResponse *SetParameterValuesResponse `xml:"SetParameterValuesResponse,omitempty"`
		GetParameterValuesResponse *GetParameterValuesResponse `xml:"GetParameterValuesResponse,omitempty"`
		GetParameterNamesResponse  *GetParameterNamesResponse  `xml:"GetParameterNamesResponse,omitempty"`
	} `xml:"Body"`
}

type TransferComplete struct {
	XMLName     xml.Name `xml:"TransferComplete"`
	CommandKey  string   `xml:"CommandKey"`
	FaultStruct struct {
		XMLName     xml.Name `xml:"FaultStruct"`
		FaultCode   int      `xml:"FaultCode"`
		FaultString string   `xml:"FaultString"`
	} `xml:"FaultStruct"`
	StartTime    CWMPTime `xml:"StartTime"`
	CompleteTime CWMPTime `xml:"CompleteTime"`
}

type RequestDownload struct {
	XMLName        xml.Name `xml:"RequestDownload"`
	FileType       string   `xml:"FileType"` // "1 Firmware Upgrade Image", "2 Web Content", etc.
	FileSize       int64    `xml:"FileSize"` // Size in bytes
	TargetFileName string   `xml:"TargetFileName"`
}
type DownloadResponse struct {
	XMLName      xml.Name `xml:"DownloadResponse"`
	Status       int      `xml:"Status"` // 0 for success, other values for errors
	StartTime    CWMPTime `xml:"StartTime"`
	CompleteTime CWMPTime `xml:"CompleteTime"`
}

type AutonomousTransferComplete struct {
	XMLName     xml.Name `xml:"AutonomousTransferComplete"`
	AnnounceURL string   `xml:"AnnounceURL,omitempty"` // Optional URL for the CPE to announce the transfer
	TransferURL string   `xml:"TransferURL,omitempty"` // Optional URL for the CPE to download the file
	FaultStruct struct {
		XMLName     xml.Name `xml:"Fault"`
		FaultCode   int      `xml:"FaultCode"`
		FaultString string   `xml:"FaultString"`
		FaultDetail struct {
			XMLName     xml.Name `xml:"FaultDetail"`
			FaultCode   int      `xml:"FaultCode"`
			FaultString string   `xml:"FaultString"`
		} `xml:"FaultDetail"`
	} `xml:"Fault"`
	FileSize int64 `xml:"FileSize"`
}
type ScheduleInform struct {
	XMLName      xml.Name `xml:"ScheduleInform"`
	DelaySeconds int      `xml:"DelaySeconds"`
	CommandKey   string   `xml:"CommandKey"`
}
type SetVouchers struct {
	XMLName     xml.Name `xml:"SetVouchers"`
	VoucherList struct {
		Vouchers []string `xml:"string"`
	} `xml:"VoucherList"`
}
type GetOptions struct {
	XMLName    xml.Name `xml:"GetOptions"`
	OptionName string   `xml:"OptionName"`
}
type Inform struct {
	XMLName  xml.Name `xml:"Inform"`
	DeviceID struct {
		XMLName      xml.Name `xml:"DeviceId"`
		Manufacturer string   `xml:"Manufacturer"`
		OUI          string   `xml:"OUI"`
		ProductClass string   `xml:"ProductClass"`
		SerialNumber string   `xml:"SerialNumber"`
	} `xml:"DeviceId"`
	ID    string `xml:"ID"`
	Event *struct {
		XMLName xml.Name `xml:"Event"`
		Events  []struct {
			XMLName    xml.Name `xml:"EventStruct"`
			EventCode  string   `xml:"EventCode"`
			CommandKey string   `xml:"CommandKey"`
			/// add field that can hold anything we missed in this xml tag children
			Content string `xml:",innerxml"`
		} `xml:"EventStruct"`
	} `xml:"Event"`
	CurrentTime   string        `xml:"CurrentTime"`
	MaxEnvelopes  int           `xml:"MaxEnvelopes"`
	RetryCount    int           `xml:"RetryCount"`
	ParameterList ParameterList `xml:"ParameterList"`
}

// ParameterList represents the list of parameters
type ParameterList struct {
	XMLName    xml.Name               `xml:"ParameterList"`
	ArrayType  string                 `xml:"http://schemas.xmlsoap.org/soap/envelope/ arrayType,attr"`
	Parameters []ParameterValueStruct `xml:"ParameterValueStruct"`
}

// ParameterValueStruct represents a single parameter value pair
type ParameterValueStruct struct {
	XMLName xml.Name `xml:"ParameterValueStruct"`
	Name    string   `xml:"Name"`
	Value   Value    `xml:"Value"`
}

// Value represents the value with its xsi:type attribute
type Value struct {
	XMLName xml.Name `xml:"Value"`
	Type    string   `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Content string   `xml:",chardata"`
}
type GetRPCMethodsResponse struct {
	XMLName    xml.Name `xml:"GetRPCMethodsResponse"`
	MethodList []string `xml:"MethodList>string"`
}

type GetParameterValuesResponse struct {
	XMLName       xml.Name      `xml:"GetParameterValuesResponse"`
	ParameterList ParameterList `xml:"ParameterList"`
}
type SetParameterValuesResponse struct {
	XMLName      xml.Name `xml:"SetParameterValuesResponse"`
	Status       int      `xml:"Status"`
	Content      string   `xml:",innerxml"` // Holds any additional content
	ParameterKey string   `xml:"ParameterKey"`
}
type GetParameterNamesResponse struct {
	XMLName       xml.Name `xml:"GetParameterNamesResponse"`
	ParameterPath string   `xml:"ParameterPath"`
	NextLevel     int      `xml:"NextLevel"` // true to get all levels, false for immediate level only
	ParameterList []struct {
		XMLName  xml.Name `xml:"ParameterInfoStruct"`
		Name     string   `xml:"Name"`
		Writable bool     `xml:"Writable"`
	} `xml:"ParameterList>ParameterInfoStruct"`
}
type Fault struct {
	XMLName     xml.Name `xml:"Fault"`
	FaultCode   string   `xml:"faultcode"`
	FaultString string   `xml:"faultstring"`
	Detail      *struct {
		XMLName     xml.Name `xml:"detail"`
		FaultCode   int      `xml:"Fault>FaultCode"`
		FaultString string   `xml:"Fault>FaultString"`
	} `xml:"detail,omitempty"`
}

// Envelope represents a SOAP envelope
