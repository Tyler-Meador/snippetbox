package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

func (app *application) serverError(response http.ResponseWriter, request *http.Request, err error) {
	var (
		method = request.Method
		uri    = request.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), slog.String("method", method), slog.String("uri", uri))

	if app.debug {
		body := fmt.Sprintf("%s\n%s", err, trace)
		http.Error(response, body, http.StatusInternalServerError)
		return
	}

	http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(response http.ResponseWriter, status int) {
	http.Error(response, http.StatusText(status), status)
}

func (app *application) render(response http.ResponseWriter, request *http.Request, status int, page string, data templateData) {
	tmps, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(response, request, err)
		return
	}

	buf := new(bytes.Buffer)

	err := tmps.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(response, request, err)
		return
	}

	response.WriteHeader(status)

	buf.WriteTo(response)
}

func (app *application) newTemplateData(request *http.Request) templateData {
	return templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(request.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(request),
		CSRFToken:       nosurf.Token(request),
	}
}

func (app *application) decodePostForm(request *http.Request, destination any) error {
	err := request.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(destination, request.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

func (app *application) isAuthenticated(request *http.Request) bool {
	isAuthenticated, ok := request.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}
