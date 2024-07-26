package xray

import (
	"bytes"

	"x-ui/logger"
	"x-ui/util/json_util"
)

type InboundConfig struct {
	Listen         json_util.RawMessage `json:"listen"` // listen cannot be an empty string
	Port           int                  `json:"port"`
	Protocol       string               `json:"protocol"`
	Settings       json_util.RawMessage `json:"settings"`
	StreamSettings json_util.RawMessage `json:"streamSettings"`
	Tag            string               `json:"tag"`
	Sniffing       json_util.RawMessage `json:"sniffing"`
}

func (c *InboundConfig) Equals(other *InboundConfig) bool {
	logger.Error(other)
	if !bytes.Equal(c.Listen, other.Listen) {
		logger.Error("pp1")
		return false
	}
	if c.Port != other.Port {
		logger.Error("pp2")
		return false
	}
	if c.Protocol != other.Protocol {
		logger.Error("pp3")
		return false
	}
	if !bytes.Equal(c.Settings, other.Settings) {
		logger.Error("pp4")
		return false
	}
	if !bytes.Equal(c.StreamSettings, other.StreamSettings) {
		logger.Error("pp5")
		return false
	}
	if c.Tag != other.Tag {
		logger.Error("pp6")
		return false
	}
	if !bytes.Equal(c.Sniffing, other.Sniffing) {
		logger.Error("pp7")
		return false
	}
	return true
}
