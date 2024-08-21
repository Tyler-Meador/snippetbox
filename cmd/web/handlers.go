package main

import (
	"errors"
	"fmt"

	"net/http"
	"strconv"

	"github.com/Tyler-Meador/snippetbox/internal/models"
	"github.com/Tyler-Meador/snippetbox/internal/validator"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type changePasswordForm struct {
	CurrentPassword     string `form:"currentPassword"`
	NewPassword         string `form:"newPassword"`
	ConfirmNewPassword  string `form:"confirmNewPassword"`
	validator.Validator `form:"-"`
}

func ping(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("OK"))
}

func (app *application) home(response http.ResponseWriter, request *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(response, request, err)
		return
	}

	data := app.newTemplateData(request)
	data.Snippets = snippets

	app.render(response, request, http.StatusOK, "home.html", data)
}

func (app *application) snippetView(response http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(request.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(response, request)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(response, request)
		} else {
			app.serverError(response, request, err)
		}

		return
	}

	data := app.newTemplateData(request)
	data.Snippet = snippet

	app.render(response, request, http.StatusOK, "view.html", data)
}

func (app *application) snippetCreate(response http.ResponseWriter, request *http.Request) {
	data := app.newTemplateData(request)

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(response, request, http.StatusOK, "create.html", data)
}

func (app *application) snippetCreatePost(response http.ResponseWriter, request *http.Request) {
	var form snippetCreateForm

	err := app.decodePostForm(request, &form)
	if err != nil {
		app.clientError(response, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(request)
		data.Form = form
		app.render(response, request, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(response, request, err)
		return
	}

	app.sessionManager.Put(request.Context(), "flash", "Snippet successfully created!")

	http.Redirect(response, request, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) userSignup(response http.ResponseWriter, request *http.Request) {
	data := app.newTemplateData(request)
	data.Form = userSignupForm{}
	app.render(response, request, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(response http.ResponseWriter, request *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(request, &form)
	if err != nil {
		app.clientError(response, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters")

	if !form.Valid() {
		data := app.newTemplateData(request)
		data.Form = form
		app.render(response, request, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(request)
			data.Form = form
			app.render(response, request, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(response, request, err)
		}

		return
	}

	app.sessionManager.Put(request.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(response, request, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(response http.ResponseWriter, request *http.Request) {
	data := app.newTemplateData(request)
	data.Form = userLoginForm{}
	app.render(response, request, http.StatusOK, "login.html", data)
}

func (app *application) userLoginPost(response http.ResponseWriter, request *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(request, &form)
	if err != nil {
		app.clientError(response, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(request)
		data.Form = form
		app.render(response, request, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(request)
			data.Form = form
			app.render(response, request, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(response, request, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(request.Context())
	if err != nil {
		app.serverError(response, request, err)
		return
	}

	app.sessionManager.Put(request.Context(), "authenticatedUserId", id)

	path := app.sessionManager.PopString(request.Context(), "redirectPathAfterLogin")
	if path != "" {
		http.Redirect(response, request, path, http.StatusSeeOther)
		return
	}

	http.Redirect(response, request, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(response http.ResponseWriter, request *http.Request) {
	err := app.sessionManager.RenewToken(request.Context())
	if err != nil {
		app.serverError(response, request, err)
		return
	}

	app.sessionManager.Remove(request.Context(), "authenticatedUserId")

	app.sessionManager.Put(request.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(response, request, "/", http.StatusSeeOther)
}

func (app *application) about(response http.ResponseWriter, request *http.Request) {
	data := app.newTemplateData(request)
	app.render(response, request, http.StatusOK, "about.html", data)
}

func (app *application) accountView(response http.ResponseWriter, request *http.Request) {
	id := app.sessionManager.GetInt(request.Context(), "authenticatedUserId")

	user, err := app.users.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(response, request, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(response, request, err)
		}
		return
	}

	data := app.newTemplateData(request)
	data.User = user

	app.render(response, request, http.StatusOK, "account.html", data)
}

func (app *application) accountPasswordUpdate(response http.ResponseWriter, request *http.Request) {
	data := app.newTemplateData(request)
	data.Form = changePasswordForm{}
	app.render(response, request, http.StatusOK, "password.html", data)
}

func (app *application) accountPasswordUpdatePost(response http.ResponseWriter, request *http.Request) {
	var form changePasswordForm

	err := app.decodePostForm(request, &form)
	if err != nil {
		app.clientError(response, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "This field must be at least 8 characters")
	form.CheckField(validator.NotBlank(form.ConfirmNewPassword), "confirmNewPassword", "This field cannot be blank")
	form.CheckField(form.ConfirmNewPassword == form.NewPassword, "confirmNewPassword", "Passwords do not match")

	if !form.Valid() {
		data := app.newTemplateData(request)
		data.Form = form
		app.render(response, request, http.StatusUnprocessableEntity, "password.html", data)
		return
	}

	err = app.users.PasswordUpdate(app.sessionManager.GetInt(request.Context(), "authenticatedUserID"), form.CurrentPassword, form.NewPassword)
	if errors.Is(err, models.ErrInvalidCredentials) {
		form.AddFieldError("currentPassword", "Current password is incorrect")

		data := app.newTemplateData(request)
		data.Form = form

		app.render(response, request, http.StatusUnprocessableEntity, "password.html", data)
	} else {
		app.serverError(response, request, err)
	}

	app.sessionManager.Put(request.Context(), "flash", "Password has been successfully changed!")

	http.Redirect(response, request, "/account/view", http.StatusSeeOther)
}
