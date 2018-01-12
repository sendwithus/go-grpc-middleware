// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_opentracing

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	defaultOptions = &options{
		filterOutFunc: nil,
		tracer:        nil,
	}
)

// FilterFunc allows users to provide a function that filters out certain methods from being traced.
//
// If it returns false, the given request will not be traced.
type FilterFunc func(ctx context.Context, fullMethodName string) bool

type options struct {
	filterOutFunc FilterFunc
	tracer        opentracing.Tracer
	// ErrorCodes are a map of grpc error codes. If true, this code will not mark the span as errored.
	errorCodes map[codes.Code]bool
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	if optCopy.tracer == nil {
		optCopy.tracer = opentracing.GlobalTracer()
	}
	return optCopy
}

type Option func(*options)

// WithFilterFunc customizes the function used for deciding whether a given call is traced or not.
func WithFilterFunc(f FilterFunc) Option {
	return func(o *options) {
		o.filterOutFunc = f
	}
}

// WithTracer sets a custom tracer to be used for this middleware, otherwise the opentracing.GlobalTracer is used.
func WithTracer(tracer opentracing.Tracer) Option {
	return func(o *options) {
		o.tracer = tracer
	}
}

// WithWhitelistedErrorCodes sets whitelisted error codes for the middleware.
func WithWhitelistedErrorCodes(cs ...codes.Code) Option {
	return func(o *options) {
		for _, c := range cs {
			if o.errorCodes == nil {
				o.errorCodes = map[codes.Code]bool{}
			}
			o.errorCodes[c] = true
		}
	}
}

// shouldMarkWithError reads the options from the list and returns whether the error should mark the span as errored
// it uses a set of whitelisted grpc error codes for this.
func shouldMarkWithError(o *options, err error) bool {
	if o == nil {
		return true
	}
	stat, statusExists := status.FromError(err)
	var whitelisted bool
	if statusExists {
		whitelisted = o.errorCodes[stat.Code()]
	}
	return !whitelisted
}
