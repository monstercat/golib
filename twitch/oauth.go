package twitch

import (
	"net/url"
	"strings"
)

type Scope string

const (
	OAuthUrl = "https://id.twitch.tv/oauth2/authorize"

	ScopeAnalyticsReadExtensions  Scope = "analytics:read:extensions"
	ScopeAnalyticsReadGames       Scope = "analytics:read:games"
	ScopeBitsRead                 Scope = "bits:read"
	ScopeChannelEditCommercial    Scope = "channel:edit:commercial"
	ScopeChannelManageBroadcast   Scope = "channel:manage:broadcast"
	ScopeChannelManageExtensions  Scope = "channel:manage:extensions"
	ScopeChannelManagePolls       Scope = "channel:manage:polls"
	ScopeChannelManagePredictions Scope = "channel:manage:predictions"
	ScopeChannelManageRedemptions Scope = "channel:manage:redemptions"
	ScopeChannelManageSchedule    Scope = "channel:manage:schedule"
	ScopeChannelManageVideos      Scope = "channel:manage:videos"
	ScopeChannelReadEditors       Scope = "channel:read:editors"
	ScopeChannelReadHypeTrain     Scope = "channel:read:hype_train"
	ScopeChannelReadPolls         Scope = "channel:read:polls"
	ScopeChannelReadPredictions   Scope = "channel:read:predictions"
	ScopeChannelReadStreamKey     Scope = "channel:read:stream_key"
	ScopeChannelReadSubscriptions Scope = "channel:read:subscriptions"
	ScopeClipsEdit                Scope = "clips:edit"
	ScopeModerationRead           Scope = "moderation:read"
	ScopeModeratorManageAutomod   Scope = "moderator:manage:automod"
	ScopeUserEdit                 Scope = "user:edit"
	ScopeUserEditFollows          Scope = "user:edit:follows"
	ScopeUserManageBlockedUsers   Scope = "user:manage:blocked_users"
	ScopeUserReadBroadcast        Scope = "user:read:broadcast"
	ScopeUserReadEmail            Scope = "user:read:email"
	ScopeUserReadFollows          Scope = "user:read:follows"
	ScopeUserReadSubscriptions    Scope = "user:read:subscriptions"
)

func (s Scope) IsValid() bool {
	switch s {
	case ScopeAnalyticsReadExtensions, ScopeAnalyticsReadGames, ScopeBitsRead, ScopeChannelEditCommercial,
		ScopeChannelManageBroadcast, ScopeChannelManageExtensions, ScopeChannelManagePolls, ScopeChannelManagePredictions,
		ScopeChannelManageRedemptions, ScopeChannelManageSchedule, ScopeChannelManageVideos, ScopeChannelReadEditors,
		ScopeChannelReadHypeTrain, ScopeChannelReadPolls, ScopeChannelReadPredictions, ScopeChannelReadStreamKey,
		ScopeChannelReadSubscriptions, ScopeClipsEdit, ScopeModerationRead, ScopeModeratorManageAutomod, ScopeUserEdit,
		ScopeUserEditFollows, ScopeUserManageBlockedUsers, ScopeUserReadBroadcast, ScopeUserReadEmail, ScopeUserReadFollows,
		ScopeUserReadSubscriptions:
		return true
	}
	return false
}

type OAuthOptions struct {
	ClientId    string
	RedirectUri string
	Scopes      []Scope
	ForceVerify bool
	State       string
}

func (o OAuthOptions) Validate() error {
	if o.ClientId == "" {
		return ErrMissingClientId
	}
	if o.RedirectUri == "" {
		return ErrMissingRedirectUri
	}
	if len(o.Scopes) == 0 {
		return ErrMissingScopes
	}
	for _, s := range o.Scopes {
		if !s.IsValid() {
			return ErrInvalidScope
		}
	}
	return nil
}

func (o OAuthOptions) convertScopes() []string {
	if len(o.Scopes) == 0 {
		return []string{}
	}
	s := make([]string, 0, len(o.Scopes))
	for _, scope := range o.Scopes {
		s = append(s, string(scope))
	}
	return s
}

func (o OAuthOptions) Values() url.Values {
	v := url.Values{}
	v.Add("client_id", o.ClientId)
	v.Add("redirect_uri", o.RedirectUri)
	v.Add("response_type", "code")
	v.Add("scope", strings.Join(o.convertScopes(), " "))
	if o.ForceVerify {
		v.Add("force_verify", "true")
	}
	if o.State != "" {
		v.Add("state", o.State)
	}
	return v
}

func GenerateOAuthUrl(options OAuthOptions) (*url.URL, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	vals := options.Values()
	return url.Parse(OAuthUrl + "?" + vals.Encode())
}

