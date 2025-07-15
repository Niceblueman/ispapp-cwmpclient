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
	h.logger.Infof("Handling SOAP response: %s", resp.XMLName.Local)
	switch {
	case resp.Body.GetRPCMethods != nil:
		return h.handleGetRPCMethods(resp.Body.GetRPCMethods)
	case resp.Body.GetParameterValues != nil:
		return h.handleGetParameterValues(resp.Body.GetParameterValues)
	case resp.Body.SetParameterValues != nil:
		return h.handleSetParameterValues(resp.Body.SetParameterValues)
	case resp.Body.Download != nil:
		return h.handleDownload(resp.Body.Download)
	case resp.Body.Reboot != nil:
		return h.handleReboot(resp.Body.Reboot)
	case resp.Body.FactoryReset != nil:
		return h.handleFactoryReset(resp.Body.FactoryReset)
	case resp.Body.AddObject != nil:
		return h.handleAddObject(resp.Body.AddObject)
	case resp.Body.DeleteObject != nil:
		return h.handleDeleteObject(resp.Body.DeleteObject)
	case resp.Body.InformResponse != nil:
		return h.handleInformResponse(resp.Body.InformResponse)
	case resp.Body.RequestXCommand != nil:
		return h.handleRequestXCommand(resp.Body.RequestXCommand)
	case resp.Body.TransferCompleteResponse != nil:
		return h.handleTransferCompleteResponse(resp.Body.TransferCompleteResponse)
	case resp.Body.RequestDownloadResponse != nil:
		return h.handleRequestDownloadResponse(resp.Body.RequestDownloadResponse)
	case resp.Body.Fault != nil:
		return h.handleFault(resp.Body.Fault)
	case resp.Body.GetParameterNames != nil:
		return h.handleGetParameterNames(resp.Body.GetParameterNames)
	// case resp.Body.SetParameterAttributes != nil:
	// 	return h.handleSetParameterAttributes(resp.Body.SetParameterAttributes)
	// case resp.Body.GetParameterAttributes != nil:
	// 	return h.handleGetParameterAttributes(resp.Body.GetParameterAttributes)
	default:
		h.logger.Warnf("Unhandled SOAP response type: %s", resp.XMLName.Local)
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
	h.logger.Info("Handling InformResponse request")
	// Implement logic to handle InformResponse
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
