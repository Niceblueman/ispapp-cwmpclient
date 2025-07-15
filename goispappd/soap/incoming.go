package soap

import (
	"encoding/xml"

	"github.com/sirupsen/logrus"
)

// ResponceEnvelope represents incoming messages from ACS to CPE
type ResponceEnvelope struct {
	XMLName *xml.Name `xml:"soap-env:Envelope"`
	Logger  *logrus.Logger
	SoapEnv *string `xml:"xmlns:soap-env,attr"`
	Cwmp    *string `xml:"xmlns:cwmp,attr"`
	Header  *struct {
		XMLName *xml.Name `xml:"soap-env:Header"`
		ID      struct {
			XMLName        *xml.Name `xml:"cwmp:ID"`
			MustUnderstand *string   `xml:"soap-env:mustUnderstand,attr"`
			Value          *string   `xml:",chardata"`
		} `xml:"cwmp:ID"`
	} `xml:"soap-env:Header"`
	Body *struct {
		// ACS-initiated RPC methods
		XMLName                  *xml.Name                 `xml:"soap-env:Body"`
		GetRPCMethods            *GetRPCMethods            `xml:"cwmp:GetRPCMethods"`
		GetParameterValues       *GetParameterValues       `xml:"cwmp:GetParameterValues"`
		SetParameterValues       *SetParameterValues       `xml:"cwmp:SetParameterValues"`
		Download                 *Download                 `xml:"cwmp:Download"`
		GetParameterNames        *GetParameterNames        `xml:"cwmp:GetParameterNames"`
		Reboot                   *Reboot                   `xml:"cwmp:Reboot"`
		FactoryReset             *FactoryReset             `xml:"cwmp:FactoryReset"`
		AddObject                *AddObject                `xml:"cwmp:AddObject"`
		DeleteObject             *DeleteObject             `xml:"cwmp:DeleteObject"`
		InformResponse           *InformResponse           `xml:"cwmp:InformResponse"`
		RequestXCommand          *RequestXCommand          `xml:"cwmp:RequestX_Command,omitempty"`
		TransferCompleteResponse *TransferCompleteResponse `xml:"cwmp:TransferCompleteResponse"`
		RequestDownloadResponse  *RequestDownloadResponse  `xml:"cwmp:RequestDownloadResponse"`
		Fault                    *FaultResponse            `xml:"cwmp:Fault,omitempty"`
	} `xml:"soap-env:Body"`
}

// ACS-initiated RPC Methods -------------------------------------------------

type GetRPCMethods struct {
	XMLName xml.Name `xml:"cwmp:GetRPCMethods"`
}

type GetParameterValues struct {
	XMLName        xml.Name       `xml:"cwmp:GetParameterValues"`
	ParameterNames ParameterNames `xml:"ParameterNames"`
}

type ParameterNames struct {
	XMLName   xml.Name `xml:"ParameterNames"`
	ArrayType string   `xml:"soap-env:arrayType,attr"`
	Names     []string `xml:"string"`
}
type SetParameterValues struct {
	XMLName       xml.Name `xml:"cwmp:SetParameterValues"`
	ParameterList struct {
		Params []struct {
			Name  string `xml:"cwmp:Name"`
			Value string `xml:"cwmp:Value"`
		} `xml:"cwmp:ParameterValueStruct"`
	} `xml:"cwmp:ParameterList"`
	ParameterKey string `xml:"cwmp:ParameterKey"` // Used for atomic commits
}

type Download struct {
	XMLName        xml.Name `xml:"cwmp:Download"`
	CommandKey     string   `xml:"cwmp:CommandKey"`
	FileType       string   `xml:"cwmp:FileType"`
	Status         int      `xml:"cwmp:Status"`
	URL            string   `xml:"cwmp:URL"`
	Username       *string  `xml:"cwmp:Username"`
	Password       *string  `xml:"cwmp:Password"`
	FileSize       *int64   `xml:"cwmp:FileSize"`
	TargetFileName string   `xml:"cwmp:TargetFileName"`
	SuccessURL     string   `xml:"cwmp:SuccessURL,omitempty"` // Optional URL for success notification
	FailureURL     string   `xml:"cwmp:FailureURL,omitempty"` // Optional URL for failure notification
	DelaySeconds   int      `xml:"cwmp:DelaySeconds"`
}

// Response Structs (ACS replies to CPE) -------------------------------------

type InformResponse struct {
	XMLName      xml.Name `xml:"cwmp:InformResponse"`
	MaxEnvelopes int      `xml:"cwmp:MaxEnvelopes"`
}

type TransferCompleteResponse struct {
	XMLName xml.Name `xml:"cwmp:TransferCompleteResponse"`
}

type RequestDownloadResponse struct {
	XMLName     xml.Name `xml:"cwmp:RequestDownloadResponse"`
	DownloadURL string   `xml:"cwmp:DownloadURL"`
}

// Common Types --------------------------------------------------------------

// GetParameterNames - ACS requests parameter names from CPE
type GetParameterNames struct {
	XMLName        xml.Name `xml:"cwmp:GetParameterNames"`
	ParameterPath  string   `xml:"cwmp:ParameterPath,omitempty"`         // e.g. "InternetGatewayDevice."
	NextLevel      int      `xml:"cwmp:NextLevel,omitempty"`             // true for next level, false for current
	ParameterNames []string `xml:"cwmp:ParameterNames>string,omitempty"` // e.g. "InternetGatewayDevice."
}

// Reboot - ACS commands the CPE to reboot
type Reboot struct {
	XMLName    xml.Name `xml:"cwmp:Reboot"`
	CommandKey string   `xml:"cwmp:CommandKey,omitempty"` // Identifier for tracking
}

// FactoryReset - ACS commands the CPE to reset to factory defaults
type FactoryReset struct {
	XMLName    xml.Name `xml:"cwmp:FactoryReset"`
	CommandKey string   `xml:"cwmp:CommandKey,omitempty"` // Identifier for tracking
}

// AddObject - ACS requests creation of a new object instance
type AddObject struct {
	XMLName      xml.Name `xml:"cwmp:AddObject"`
	ObjectName   string   `xml:"cwmp:ObjectName"`             // e.g. "InternetGatewayDevice.LANDevice.1"
	ParameterKey string   `xml:"cwmp:ParameterKey,omitempty"` // Used for atomic operations
}

// DeleteObject - ACS requests deletion of an object instance
type DeleteObject struct {
	XMLName      xml.Name `xml:"cwmp:DeleteObject"`
	ObjectName   string   `xml:"cwmp:ObjectName"`             // e.g. "InternetGatewayDevice.LANDevice.1"
	ParameterKey string   `xml:"cwmp:ParameterKey,omitempty"` // Used for atomic operations
}

type RebootResponse struct {
	XMLName xml.Name `xml:"cwmp:RebootResponse"`
}

type FactoryResetResponse struct {
	XMLName xml.Name `xml:"cwmp:FactoryResetResponse"`
}

type AddObjectResponse struct {
	XMLName        xml.Name `xml:"cwmp:AddObjectResponse"`
	InstanceNumber int      `xml:"cwmp:InstanceNumber"` // The new instance number created
	Status         int      `xml:"cwmp:Status"`         // 0 = success, 1 = error
}

type DeleteObjectResponse struct {
	XMLName xml.Name `xml:"cwmp:DeleteObjectResponse"`
	Status  int      `xml:"cwmp:Status"` // 0 = success, 1 = error
}

// Supporting struct
type ParameterInfoStruct struct {
	Name     string `xml:"cwmp:Name"`
	Writable bool   `xml:"cwmp:Writable"`
}

type FaultResponse struct {
	XMLName     xml.Name `xml:"cwmp:Fault"`
	FaultCode   string   `xml:"cwmp:FaultCode"`
	FaultString string   `xml:"cwmp:FaultString"`
	FaultDetail struct {
		XMLName     xml.Name `xml:"cwmp:FaultDetail"`
		FaultCode   string   `xml:"cwmp:FaultCode"`
		FaultString string   `xml:"cwmp:FaultString"`
	} `xml:"cwmp:FaultDetail,omitempty"`
}

type RequestXCommand struct {
	XMLName    xml.Name `xml:"cwmp:X_Command"`
	CommandKey string   `xml:"cwmp:CommandKey"`
	Parameters struct {
		XMLName xml.Name `xml:"cwmp:Parameters"`
		Command string   `xml:"cwmp:Command"`
	}
}
