// Goro
//
// Created by Posse in NYC
// http://goposse.com
//
// Copyright (c) 2016 Posse Productions LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

// ContextKey - type used to store goro defined context keys
type ContextKey int

// PathContextKey - contains the original path
const PathContextKey ContextKey = 0

// ParametersContextKey - contains the parsed parameter keys and values
const ParametersContextKey ContextKey = 1

// CatchAllValueContextKey - contains the found catch all value
const CatchAllValueContextKey ContextKey = 2

// ErrorValueContextKey - contains the value of any generic error that occurred
const ErrorValueContextKey ContextKey = 3

// error codes

// RouterErrorCode - type used to represent an error in the routing request
type RouterErrorCode int

// ErrorCodePanic - error code used when recovering from a panic
const ErrorCodePanic RouterErrorCode = 777
