package cwmp

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Niceblueman/goispappd/device"
	"github.com/Niceblueman/goispappd/internal/config"
	"github.com/Niceblueman/goispappd/soap"
	"github.com/sirupsen/logrus"
)

// CWMPClient manages the TR-069 client state
type CWMPClient struct {
	config     *config.Configuration
	httpClient *http.Client
	logger     *logrus.Logger
	Handler    *Handler
	dataModel  *device.Device
	Response   *soap.ResponceEnvelope
}

// NewCWMPClient initializes a new CWMP client
func NewCWMPClient(config *config.Configuration, logger *logrus.Logger) *CWMPClient {
	return &CWMPClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
		dataModel:  &device.Device{},
		Handler:    NewHandler(logger),
		Response:   soap.NewResponceEnvelope(logger),
	}
}

// Initialize sets up the client and loads initial data
func (c *CWMPClient) Initialize(ctx context.Context) error {
	c.logger.Info("Initializing CWMP client")
	go c.periodicInform(ctx)
	return nil
}

func (c *CWMPClient) SendEnvelope(envelope *soap.RequestEnvelope) error {
	c.logger.Infof("Sending SOAP envelope: %s", envelope.XMLName.Local)
	buf, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.ACSURL, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "urn:dslforum-org:cwmp-1-2")
	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SOAP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK status: %s", resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if err := c.Response.Load(respBody); err != nil {
		return fmt.Errorf("failed to load SOAP response: %w", err)
	}
	return c.Handler.HandleResponse(c.Response)
}

// periodicInform sends periodic Inform messages to the ACS
func (c *CWMPClient) periodicInform(ctx context.Context) {
	ticker := time.NewTicker(c.config.PeriodicInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping periodic inform")
			return
		case <-ticker.C:
			// Pause ticker while waiting for handler
			ticker.Stop()
			informDone := make(chan error, 1)
			go func() {
				informDone <- c.SendInform("2 PERIODIC")
			}()
			timeout := 30 * time.Second
			select {
			case err := <-informDone:
				if err != nil {
					c.logger.Errorf("Failed to send periodic inform: %v", err)
				}
				// Resume ticker
				ticker = time.NewTicker(c.config.PeriodicInterval)
			case <-time.After(timeout):
				c.logger.Errorf("Periodic inform handler timeout, ending session and increasing interval by 5s")
				c.config.PeriodicInterval += 5 * time.Second
				return
			}
		}
	}
}

// SendInform constructs and sends an Inform message
func (c *CWMPClient) SendInform(eventCode string) error {
	envelope := soap.NewRequestEnvelope()
	envelope.LoadInformRequest()
	body, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Inform XML: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.ACSURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "urn:dslforum-org:cwmp-1-2")
	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Inform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK status: %s", resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if err := c.Response.Load(respBody); err != nil {
		return fmt.Errorf("failed to load SOAP response: %w", err)
	}
	if err := c.Handler.HandleResponse(c.Response); err != nil {
		return fmt.Errorf("failed to handle SOAP response: %w", err)
	}
	return nil
}
