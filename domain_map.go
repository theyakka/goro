// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2019 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const DomainMapNakedSubdomainKey = ":naked"
const DomainMapWildcardSubdomainKey = ":wildcard"

var domainCheckRegexp *regexp.Regexp

// DomainMap - maps (sub)domains to routers
type DomainMap struct {
	domains              []string
	subdomains           []string
	routerMap            map[string]*Router
	subdomainHostMatches map[string]string
	hasWildcard          bool

	// NotFoundHandler - if the (sub)domain is not mapped, call this handler
	NotFoundHandler ContextHandlerFunc
}

// NewDomainMap - creates a new domain map for the provided domains
func NewDomainMap(domains ...string) *DomainMap {
	joinedDomains := strings.Join(domains, "|")
	domainCheckPattern := fmt.Sprintf(`^((?:[a-z0-9]{0,69}\.)*)(?:%s)(?::\d+)?$`, joinedDomains)
	regex, regexErr := regexp.Compile(domainCheckPattern)
	if regexErr != nil {
		// TODO ..
	}
	domainCheckRegexp = regex
	return &DomainMap{
		domains:              domains,
		subdomains:           []string{},
		routerMap:            map[string]*Router{},
		subdomainHostMatches: map[string]string{},
		hasWildcard:          false,
	}
}

// NewRouter - Creates a new Router, registers it in the domain map and returns it for use
func (dm *DomainMap) NewRouter(subdomainPattern string) *Router {
	router := NewRouter()
	subdomains := strings.Split(subdomainPattern, "|")
	for _, subdomain := range subdomains {
		key := subdomain
		if subdomain == "*" {
			key = DomainMapWildcardSubdomainKey
			dm.hasWildcard = true
		} else if subdomain == "<*>" {
			key = DomainMapNakedSubdomainKey
		}
		dm.AddRouter(key, router)
	}
	return router
}

// NewRouters - Creates a new Router for each of the defined subdomains and registers it
// with the DomainMap
func (dm *DomainMap) NewRouters(subdomains ...string) []*Router {
	var routers []*Router
	for _, subdomain := range subdomains {
		key := subdomain
		if subdomain == "*" {
			key = DomainMapWildcardSubdomainKey
			dm.hasWildcard = true
		} else if subdomain == "<*>" {
			key = DomainMapNakedSubdomainKey
		}
		router := NewRouter()
		dm.AddRouter(key, router)
		routers = append(routers, router)
	}
	return routers
}

// AddRouter - Register a router for a domain pattern (regex)
func (dm *DomainMap) AddRouter(subdomain string, router *Router) {
	dm.subdomains = append(dm.subdomains, subdomain)
	dm.routerMap[subdomain] = router
}

func (dm *DomainMap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var router *Router
	hostMinusPort := strings.Split(req.Host, ":")[0]
	subdomain, isMapped := dm.isMappedHost(hostMinusPort)
	if isMapped {
		router = dm.routerMap[subdomain]
	} else if dm.isNakedHost(hostMinusPort) {
		router = dm.routerMap[DomainMapNakedSubdomainKey]
	} else if dm.hasWildcard {
		router = dm.routerMap[DomainMapWildcardSubdomainKey]
	}
	if router == nil {
		if dm.NotFoundHandler != nil {
			dm.domainRouterNotFoundHandler().ServeHTTP(w, req)
		} else {
			// no handler & no matches found. fail.
			errorHandler(w, req, "Forbidden", http.StatusForbidden)
		}
		return
	}
	router.ServeHTTP(w, req)
}

func (dm DomainMap) isMappedHost(host string) (string, bool) {
	for hostString, subdomain := range dm.subdomainHostMatches {
		if hostString == host {
			return subdomain, true
		}
	}
	matches := domainCheckRegexp.FindStringSubmatch(host)
	if len(matches) <= 1 {
		return "", false // naked domain or bad domain
	}
	subdomainMatch := strings.TrimRight(matches[1], ".")
	for _, subdomain := range dm.subdomains {
		if subdomain == subdomainMatch {
			dm.subdomainHostMatches[host] = subdomain
			return subdomain, true
		}
	}
	return "", false
}

func (dm DomainMap) isNakedHost(host string) bool {
	for _, registeredHost := range dm.domains {
		if host == registeredHost {
			return true
		}
	}
	return false
}

func (dm DomainMap) domainRouterNotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO - add additional context as to why it failed
		context := NewHandlerContext(r, w, nil)
		dm.NotFoundHandler(context)
	})
}
