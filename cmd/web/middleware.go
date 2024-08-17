package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		response.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "deny")
		response.Header().Set("X-XSS-Protection", "0")
		response.Header().Set("Server", "Go")

		next.ServeHTTP(response, request)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var (
			ip     = request.RemoteAddr
			proto  = request.Proto
			method = request.Method
			uri    = request.URL.RequestURI()
		)

		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(response, request)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				response.Header().Set("Connection", "close")
				app.serverError(response, request, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(response, request)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !app.isAuthenticated(request) {
			http.Redirect(response, request, "/user/login", http.StatusSeeOther)
			return
		}

		response.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(response, request)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		id := app.sessionManager.GetInt(request.Context(), "authenticatedUserId")
		if id == 0 {
			next.ServeHTTP(response, request)
			return
		}

		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(response, request, err)
			return
		}

		if exists {
			ctx := context.WithValue(request.Context(), isAuthenticatedContextKey, true)
			request = request.WithContext(ctx)
		}

		next.ServeHTTP(response, request)
	})
}
