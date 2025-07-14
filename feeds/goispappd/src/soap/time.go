package soap

import (
	"encoding/xml"
	"time"
)

// CWMPTime handles TR-069's ISO 8601 timestamps
type CWMPTime struct {
	time.Time
}

func (c *CWMPTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(c.Time.Format(time.RFC3339), start)
}

func (c *CWMPTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	c.Time = parsed
	return nil
}
