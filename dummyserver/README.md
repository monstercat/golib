# Dummy Server
For faking external service HTTP servers in your unit tests.

## API Should Have Configurable Host
Make sure your **configuration points to localhost** in whatever code is making external requests.

```
// TESTS: conf := {Host: "http://localhost:1234"}
// PROD:  conf := {Host: "http://somerealapi.com"}
func GetUserYouTubeChannels(conf *Config, userId string) (Result, error) {
  return MakeRequest(conf.Host+"/user/"+userId)
}
```

## Set Up Sever In Your Tests
Start up your server before tests that will make external requests.

### Example
Sets up a straight url=>json dummy route and also a custom handler.

Note that the **port** matches the one in the config when testing.

```
DummyStreamlabs := &DummyServer{
	Port:  1234,
	Debug: true,
	Routes: []*DummyRoute{
		{
			// Profiles of LegacyLarry
			Path:   "/api/v1/developer/users/" + LegacyLarryUIDStr,
			Status: http.StatusOK,
			Body:   `{"twitch":{"twitch_id":18057643,"display_name":"LegacyLarry","name":"legacylarry","partnered":0}}`,
		},
		{
			Path: "/api/v1/developer/subscriptions",
			Handler: func(w http.ResponseWriter, req *http.Request) {
				vals := req.URL.Query()
				w.Header().Add("Content-Type", "application/json")
				response := map[string]interface{}{}
				if vals.Get("user_id") == "1337" {
					response["data"] = []interface{}{
						map[string]interface{}{
							"id":            1234,
							"streamlabs_id": "1234",
							"status":        "active",
						},
					}
				} else {
					response["data"] = []interface{}{}
				}
				respBytes, err := json.Marshal(response)
				if err != nil {
					w.WriteHeader(500)
				} else {
					w.WriteHeader(200)
					w.Write(respBytes)
				}
			},
		},
	},
}
if err := DummyStreamlabs.Start(); err != nil {
	panic(err)
}
```