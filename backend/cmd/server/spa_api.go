//go:build noembed

package main

import "net/http"

func registerSPA(_ *http.ServeMux) {}
