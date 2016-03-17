package goro

import (
	"net/http"
	"regexp"
)

// DomainMap - maps (sub)domains to routers
type DomainMap struct {
	orderedPatterns    []string
	registeredPatterns map[string]*Router
	matchedHosts       map[string]*Router

	// NotFoundHandler - if the (sub)domain is not mapped, call this handler
	NotFoundHandler http.HandlerFunc
}

// NewDomainMap - creates a new domain map
func NewDomainMap() *DomainMap {
	return &DomainMap{
		registeredPatterns: make(map[string]*Router),
		matchedHosts:       make(map[string]*Router),
	}
}

// InvalidateMatchedHosts - resets any matched (sub)domains that have been cached
func (domainMap *DomainMap) InvalidateMatchedHosts() {
	domainMap.matchedHosts = make(map[string]*Router)
}

// RegisterRouter - Register a router for a domain pattern (regex)
func (domainMap *DomainMap) RegisterRouter(pattern string, router *Router) {
	domainMap.orderedPatterns = append(domainMap.orderedPatterns, pattern)
	domainMap.registeredPatterns[pattern] = router
}

func (domainMap *DomainMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check cached handlers (keyed on host)
	for host, router := range domainMap.matchedHosts {
		if r.Host == host {
			router.ServeHTTP(w, r)
			return
		}
	}
	// no cached handler found, consult the registered patterns
	for _, pattern := range domainMap.orderedPatterns {
		regex, _ := regexp.Compile(pattern)
		if regex.MatchString(r.Host) {
			router := domainMap.registeredPatterns[pattern]
			router.ServeHTTP(w, r)
			domainMap.matchedHosts[r.Host] = router
			return
		}
	}

	if domainMap.NotFoundHandler != nil {
		domainMap.NotFoundHandler(w, r)
	} else {
		// no handler & no matches found. fail.
		http.Error(w, "Forbidden", 403)
	}
}
