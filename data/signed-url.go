package data

import (
	"fmt"
)

type SignedUrlConfig struct {
	Download    bool
	Filename    string
	ContentType string
}

func (c SignedUrlConfig) GetDisposition() string {
	if !c.Download {
		return "inline"
	}
	if c.Filename == "" {
		return "attachment"
	}
	return fmt.Sprintf("attachment; filename=\"%s\"", c.Filename)
}

func (c SignedUrlConfig) GetContentType() string {
	if c.ContentType == "" {
		return "binary/octet-stream"
	}
	return c.ContentType
}
