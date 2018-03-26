// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2018 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

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
	NotFoundHandler http.Handler
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

// NewRouter - Creates a new Router, registers it in the domain map and returns it for use
func (domainMap *DomainMap) NewRouter(subdomain string) *Router {
	router := NewRouter()
	domainMap.AddRouter(subdomain, router)
	return router
}

// AddRouter - Register a router for a domain pattern (regex)
func (domainMap *DomainMap) AddRouter(subdomain string, router *Router) {
	domainMap.orderedPatterns = append(domainMap.orderedPatterns, subdomain)
	domainMap.registeredPatterns[subdomain] = router
}

func (domainMap *DomainMap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// check cached handlers (keyed on host)
	for host, router := range domainMap.matchedHosts {
		if req.Host == host {
			router.ServeHTTP(w, req)
			return
		}
	}
	// no cached handler found, consult the registered patterns
	for _, pattern := range domainMap.orderedPatterns {
		regex, _ := regexp.Compile(pattern)
		if regex.MatchString(req.Host) {
			router := domainMap.registeredPatterns[pattern]
			router.ServeHTTP(w, req)
			domainMap.matchedHosts[req.Host] = router
			return
		}
	}

	if domainMap.NotFoundHandler != nil {
		domainMap.NotFoundHandler.ServeHTTP(w, req)
	} else {
		// no handler & no matches found. fail.
		errorHandler(w, req, "Forbidden", http.StatusForbidden)
	}
}
