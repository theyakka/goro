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
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ServeFile replies to the request with the contents of the named
// file.
//
// This implementation is different to the standard library
// implementation in that it doesn't concern itself with
// directories, and does not account for redirecting paths ending in
// /index.html. If you wish to retain that functionality, you should
// set up your routes accordingly.
//
// ServeFile will only catch common errors, for example 404s or
// access related errors and wont catch lower-level errors. Due to the
// implementation details in the Go standard library (which we still
// rely on) some of those errors will fall through to the standard
// error reporting mechanisms.
func ServeFile(ctx *HandlerContext, filename string, statusCode int) {
	router := ctx.router
	req := ctx.Request
	if containsDotDotSegment(req.URL.Path) {
		// respond with an error because the url path cannot contain '..'
		router.emitError(ctx, http.StatusBadRequest, "the request path cannot contain a '..' segment", RouterContentErrorCode, nil)
		return
	}
	// try to open the file
	dirPart, filePart := filepath.Split(filename)
	dir := http.Dir(dirPart)
	file, fileErr := dir.Open(filePart)
	if fileErr != nil {
		if os.IsNotExist(fileErr) {
			router.emitError(ctx, http.StatusNotFound, "the file you requested to serve was not found", RouterContentErrorCode, fileErr)
			return
		}
		router.emitError(ctx, http.StatusInternalServerError, fileErr.Error(), RouterContentErrorCode, fileErr)
		return
	}
	// if the file close operation fails we just log the error to debug
	defer func(f http.File) {
		closeErr := file.Close()
		if closeErr != nil {
			logger.Println("Error closing file. Details =", closeErr)
		}
	}(file)
	// check the file exists
	fileInfo, statErr := file.Stat()
	if statErr != nil {
		router.emitError(ctx, http.StatusInternalServerError, statErr.Error(), RouterContentErrorCode, statErr)
		return
	}
	// serve the file
	ext := strings.ToLower(filepath.Ext(filePart))
	contentType := mime.TypeByExtension(ext)
	ctx.ResponseWriter.Header().Add("Content-Type", contentType)
	ctx.ResponseWriter.WriteHeader(statusCode)
	http.ServeContent(ctx.ResponseWriter, req, fileInfo.Name(), fileInfo.ModTime(), file)
}

// containsDotDotSegment checks to see if any part of the split path (fields) contains
// a path segment referring to the parent directory. e.g.: /my/path/../this
func containsDotDotSegment(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

// isSlashRune checks to see if a rune is a forward or backward slash
func isSlashRune(r rune) bool { return r == '/' || r == '\\' }
