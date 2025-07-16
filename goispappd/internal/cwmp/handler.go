package cwmp

import (
	"github.com/Niceblueman/goispappd/soap"
	"github.com/sirupsen/logrus"
)

var (
	METHODS = []string{
		"GetParameterValues",
		"SetParameterValues",
		"GetParameterNames",
		"SetParameterAttributes",
		"GetParameterAttributes",
		"AddObject",
		"DeleteObject",
		"Download",
		"Upload",
		"Reboot",
		"FactoryReset",
		"GetRPCMethods",
		"Inform",
		"TransferComplete",
		"AutonomousTransferComplete",
		"X_Command",
	}
)

type Handler struct {
	// Handle incoming SOAP requests
	logger *logrus.Logger
	client *CWMPClient
}

// NewHandler initializes a new CWMP handler
func NewHandler(logger *logrus.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// HandleRequest processes the incoming SOAP request
func (h *Handler) HandleResponse(resp *soap.ResponceEnvelope) error {
	METHOD := resp.GetMethodSwitch()
	switch METHOD {
	case "GetRPCMethods":
		return h.handleGetRPCMethods(resp.Body.GetRPCMethods)
	case "GetParameterValues":
		return h.handleGetParameterValues(resp.Body.GetParameterValues)
	case "SetParameterValues":
		return h.handleSetParameterValues(resp.Body.SetParameterValues)
	case "Download":
		return h.handleDownload(resp.Body.Download)
	case "Reboot":
		return h.handleReboot(resp.Body.Reboot)
	case "FactoryReset":
		return h.handleFactoryReset(resp.Body.FactoryReset)
	case "AddObject":
		return h.handleAddObject(resp.Body.AddObject)
	case "DeleteObject":
		return h.handleDeleteObject(resp.Body.DeleteObject)
	case "InformResponse":
		return h.handleInformResponse(resp.Body.InformResponse)
	case "RequestXCommand":
		return h.handleRequestXCommand(resp.Body.RequestXCommand)
	case "TransferCompleteResponse":
		return h.handleTransferCompleteResponse(resp.Body.TransferCompleteResponse)
	case "RequestDownloadResponse":
		return h.handleRequestDownloadResponse(resp.Body.RequestDownloadResponse)
	case "Fault":
		return h.handleFault(resp.Body.Fault)
	case "GetParameterNames":
		return h.handleGetParameterNames(resp.Body.GetParameterNames)
	// case resp.Body.SetParameterAttributes:
	// 	return h.handleSetParameterAttributes(resp.Body.SetParameterAttributes)
	// case resp.Body.GetParameterAttributes:
	// 	return h.handleGetParameterAttributes(resp.Body.GetParameterAttributes)
	default:
		h.logger.Warnf("Unhandled SOAP response type: %s", METHOD)
		return nil
	}
}

func (h *Handler) handleGetRPCMethods(_ *soap.GetRPCMethods) error {
	h.logger.Info("Handling GetRPCMethods request")
	// Implement logic to handle GetRPCMethods
	envelope := soap.NewRequestEnvelope()
	envelope.LoadRPCMethods()
	return h.client.SendEnvelope(envelope)
}
func (h *Handler) handleGetParameterValues(method *soap.GetParameterValues) error {
	h.logger.Info("Handling GetParameterValues request")
	// Implement logic to handle GetParameterValues
	return nil
}
func (h *Handler) handleSetParameterValues(method *soap.SetParameterValues) error {
	h.logger.Info("Handling SetParameterValues request")
	// Implement logic to handle SetParameterValues
	return nil
}
func (h *Handler) handleDownload(method *soap.Download) error {
	h.logger.Info("Handling Download request")
	// Implement logic to handle Download
	return nil
}
func (h *Handler) handleReboot(method *soap.Reboot) error {
	h.logger.Info("Handling Reboot request")
	// Implement logic to handle Reboot
	return nil
}
func (h *Handler) handleFactoryReset(method *soap.FactoryReset) error {
	h.logger.Info("Handling FactoryReset request")
	// Implement logic to handle FactoryReset
	return nil
}
func (h *Handler) handleAddObject(method *soap.AddObject) error {
	h.logger.Info("Handling AddObject request")
	// Implement logic to handle AddObject
	return nil
}
func (h *Handler) handleDeleteObject(method *soap.DeleteObject) error {
	h.logger.Info("Handling DeleteObject request")
	// Implement logic to handle DeleteObject
	return nil
}
func (h *Handler) handleInformResponse(method *soap.InformResponse) error {

	return nil
}
func (h *Handler) handleRequestXCommand(method *soap.RequestXCommand) error {
	h.logger.Info("Handling RequestXCommand request")
	// Implement logic to handle RequestXCommand
	return nil
}
func (h *Handler) handleTransferCompleteResponse(method *soap.TransferCompleteResponse) error {
	h.logger.Info("Handling TransferCompleteResponse request")
	// Implement logic to handle TransferCompleteResponse
	return nil
}
func (h *Handler) handleRequestDownloadResponse(method *soap.RequestDownloadResponse) error {
	h.logger.Info("Handling RequestDownloadResponse request")
	// Implement logic to handle RequestDownloadResponse
	return nil
}
func (h *Handler) handleFault(method *soap.FaultResponse) error {
	h.logger.Errorf("Handling Fault response: %s", method.FaultCode)
	// Implement logic to handle Fault
	return nil
}
func (h *Handler) handleTransferComplete(method *soap.TransferComplete) error {
	h.logger.Info("Handling TransferComplete request")
	// Implement logic to handle TransferComplete
	return nil
}
func (h *Handler) handleAutonomousTransferComplete(method *soap.AutonomousTransferComplete) error {
	h.logger.Info("Handling AutonomousTransferComplete request")
	// Implement logic to handle AutonomousTransferComplete
	return nil
}
func (h *Handler) handleGetParameterNames(method *soap.GetParameterNames) error {
	h.logger.Info("Handling GetParameterNames request")
	// Implement logic to handle GetParameterNames
	return nil
}
