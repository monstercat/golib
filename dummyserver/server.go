package dummyserver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	. "github.com/monstercat/golib/url"
)

type DummyServer struct {
	Port              int
	Routes            []*DummyRoute
	Running           bool
	Debug             bool
	PathQueryHandlers map[string][]*DummyRoute
}

type DummyRoute struct {
	Method      string
	Path        string // eg: /users/profiles/Vindexus
	QueryString string // eg: expired=1

	// if not nil, this will be run
	Handler func(w http.ResponseWriter, req *http.Request)

	// If no handler func is provided, these will be used
	Status  int
	Body    string
	Headers map[string]string
}

/*
type SimpleRoute struct {
	Url      string // path + query string
	Status   int
	Response interface{}
}

func (ds *DummyServer) AddSimpleRoutes(srs []interface{}) {

}
*/
func (ds *DummyServer) Log(args ...interface{}) {
	if !ds.Debug {
		return
	}

	args = append([]interface{}{"(dmysrvr)"}, args...)
	fmt.Println(args...)
}

func (ds *DummyServer) Start() error {
	if ds.Port == 0 {
		return errors.New("port can't be zero")
	}

	if err := ds.RegisterRoutes(); err != nil {
		return err
	}

	ds.Running = true

	return nil
}

func (ds *DummyServer) RegisterRoutes() error {
	if len(ds.Routes) == 0 {
		return errors.New("no routes defined")
	}

	ds.PathQueryHandlers = map[string][]*DummyRoute{}

	// For each path, we will create a hander
	// If multiple DummyRoutes share a path but have different query strings then the single handler
	// for that path will just match the query string
	for i, r := range ds.Routes {
		if r.Handler == nil && r.Status == 0 {
			return errors.New(fmt.Sprintf("Route[%d] (%s) has no handler AND no status code", i, r.Path))
		}

		if r.Handler != nil && r.QueryString != "" {
			return errors.New(fmt.Sprintf("Route[%d] (%s). You can't have both a custom handler and a querystring"))
		}

		if r.Path[:1] != "/" {
			return errors.New(fmt.Sprintf("Route[%d] (%s). Must start with /"))
		}

		if strings.Contains(r.Path, "?") {
			return errors.New(fmt.Sprintf("Route[%d] (%s). Path can't have query params", i, r.Path))
		}

		if _, ok := ds.PathQueryHandlers[r.Path]; !ok {
			ds.PathQueryHandlers[r.Path] = make([]*DummyRoute, 0)
		}

		// Let's sort the query string provided so that we always match properly
		// This is so if we declare a dummy route at "/account?a=first&b=second" it will still match a request
		// that comes in for "/account?b=second&a=first". This can happen when using maps because Go randomizes
		// the order of maps
		sorted, err := SortQueryString(r.QueryString)
		if err != nil {
			panic(err)
		}
		r.QueryString = sorted

		ds.PathQueryHandlers[r.Path] = append(ds.PathQueryHandlers[r.Path], r)
	}

	hserver := http.NewServeMux()

	for path, queries := range ds.PathQueryHandlers {
		var handler func(w http.ResponseWriter, req *http.Request)

		// For a given path with its own specific handler, we use the given handler
		if len(queries) == 1 && queries[0].Handler != nil {
			handler = queries[0].Handler
			ds.Log("Custom handler for path", path)
		} else {
			handler = func(w http.ResponseWriter, req *http.Request) {
				sortedQuery := SortedQueryString(req.URL.Query())
				var route *DummyRoute
				for _, dr := range ds.PathQueryHandlers[req.URL.Path] {
					if dr.QueryString == sortedQuery {
						route = dr
						break
					}
				}

				if route == nil {
					w.WriteHeader(500)
					// Will the loop change the value of path to the last value for ALL handlers?
					msg := fmt.Sprintf("Could not find a dummy route for path '%s' and query '%s'", path, sortedQuery)
					ds.Log(msg)
					w.Write([]byte(msg))
				} else {
					for key, value := range route.Headers {
						w.Header().Add(key, value)
					}

					// Default to JSON
					if w.Header().Get("Content-Type") == "" {
						w.Header().Add("Content-Type", "application/json")
					}

					w.Write([]byte(route.Body))
				}
			}
			ds.Log("Handler for path", path, "with", len(queries), "possible queries")
		}

		hserver.HandleFunc(path, handler)
	}

	hserver.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			ds.Log("404: ", r.URL.String())
			w.WriteHeader(http.StatusNotFound)

			return
		}
	})

	// Is this go func needed? I think it might be if you want to run two servers at once
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", ds.Port), hserver)
		ds.Log("Dummy server listening on port", ds.Port)
	}()

	return nil
}
