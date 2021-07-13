package youtube

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ListChannelsUrl = BaseUrlV3 + "/channels"
)

type ChannelPart string

const (
	ChannelPartAuditDetails        ChannelPart = "auditDetails"
	ChannelPartBrandingSettings    ChannelPart = "brandingSettings"
	ChannelPartContentDetails      ChannelPart = "contentDetails"
	ChannelPartContentOwnerDetails ChannelPart = "contentOwnerDetails"
	ChannelPartId                  ChannelPart = "id"
	ChannelPartLocalizations       ChannelPart = "localizations"
	ChannelPartSnippet             ChannelPart = "snippet"
	ChannelPartStatistics          ChannelPart = "statistics"
	ChannelPartStatus              ChannelPart = "status"
	ChannelPartTopicDetails        ChannelPart = "topicDetails"
)

func (p ChannelPart) IsValid() bool {
	switch p {
	case ChannelPartAuditDetails, ChannelPartBrandingSettings, ChannelPartContentDetails, ChannelPartContentOwnerDetails,
		ChannelPartId, ChannelPartLocalizations, ChannelPartSnippet, ChannelPartStatistics, ChannelPartStatus, ChannelPartTopicDetails:
		return true
	}
	return false
}

type Channel struct {
	Kind    string          `json:"kind"`
	Etag    string          `json:"etag"`
	Id      string          `json:"id"`
	Snippet *ChannelSnippet `json:"snippet"`
}

type ChannelSnippet struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	PublishedAt time.Time            `json:"publishedAt"`
	Thumbnails  map[string]ImageLink `json:"thumbnails"`
	// TODO: localized
}

type ListChannelsOpts struct {
	Parts []ChannelPart

	/// Filters
	ForUsername string
	Id          string
	ManagedByMe bool
	Mine        bool

	// Optional Params
	HL                     string
	MaxResults             int
	OnBehalfOfContentOwner string
	PageToken              string
}

func (o ListChannelsOpts) convertParts() []string {
	if len(o.Parts) == 0 {
		return []string{}
	}
	p := make([]string, 0, len(o.Parts))
	for _, prompt := range o.Parts {
		p = append(p, string(prompt))
	}
	return p
}

func (o ListChannelsOpts) Validate() error {
	if len(o.Parts) == 0 {
		return ErrMissingParts
	}
	for _, p := range o.Parts {
		if !p.IsValid() {
			return ErrInvalidPart
		}
	}
	return nil
}

func (o ListChannelsOpts) Values() url.Values {
	vals := url.Values{}
	vals.Add("part", strings.Join(o.convertParts(), ","))

	if o.ForUsername != "" {
		vals.Add("forUsername", o.ForUsername)
	}
	if o.Id != "" {
		vals.Add("id", o.Id)
	}
	if o.ManagedByMe {
		vals.Add("managedbyMe", "true")
	}
	if o.Mine {
		vals.Add("mine", "true")
	}
	if o.HL != "" {
		vals.Add("hl", o.HL)
	}
	if o.MaxResults > 0 {
		vals.Add("maxResults", strconv.Itoa(o.MaxResults))
	}
	if o.OnBehalfOfContentOwner != "" {
		vals.Add("onBehalfOfContentOwner", o.OnBehalfOfContentOwner)
	}
	if o.PageToken != "" {
		vals.Add("pageToken", o.PageToken)
	}
	return vals
}

// MyChannel retrieves the channel related to the provided access token.
// We will be setting Mine to true, thereby forcing the request to return only a single channel
// @see https://developers.google.com/youtube/v3/docs/channels/list
func MyChannel(accessToken string, timeout time.Duration) (*Channel, error) {
	opts := ListChannelsOpts{
		Parts: []ChannelPart{
			ChannelPartSnippet,
		},
		Mine: true,
	}

	res, err := Run(&Request{
		Method:      http.MethodGet,
		Url:         ListChannelsUrl,
		Params:      opts.Values(),
		AccessToken: accessToken,
		Timeout:     timeout,
	})
	if err != nil {
		return nil, err
	}

	// The API returns more fields, but for this case, we only want the first channel (we assume there should only
	// be one).
	var x struct {
		Items []*Channel `json:"items"`
	}
	if err := DecodeResponse(res, &x); err != nil {
		return nil, err
	}

	if len(x.Items) == 0 {
		return nil, ErrNotFound
	}
	return x.Items[0], nil
}
