package html

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// UsersMain renders the users page main content.
func UsersMain(currentUser *domain.User, users []*domain.User, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			UsersListPartial(currentUser, users, csrf),
		),
	)
}

// UsersListPartial renders users list partial.
func UsersListPartial(currentUser *domain.User, users []*domain.User, csrf string) Node {
	if len(users) == 0 {
		return Div(Class("px-3 py-4 text-center"),
			P(Text("no users found")),
		)
	}

	return Div(
		Table(Class("table-fixed w-full"),
			THead(
				Tr(
					Th(Class("text-left"), Text("Username")),
					Th(Class("text-left"), Text("Role")),
					Th(Class("text-left"), Text("Created at")),
					Th(Class("text-left"), Text("Actions")),
				),
			),
			TBody(
				Map(users, func(user *domain.User) Node {
					return Tr(
						Td(Text(user.Username)),
						Td(Text(string(user.Role))),
						Td(Text(user.GetCreatedAt().Format(time.DateTime))),
						Td(
							If(currentUser.Role == domain.Admin && currentUser.ID != user.ID,
								A(Class("hover:underline text-amber-600 font-semibold"),
									hx.Post("/delete-user"),
									hx.Confirm("Are you sure?"),
									hx.Vals(string(must(json.Marshal(map[string]string{
										"csrf":     csrf,
										"username": user.Username,
									})))),
									Href("#"),
									Text("DELETE"),
								),
							),
						),
					)
				}),
			),
		),
		Div(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"),
			Button(buttonClasses(),
				Type("button"),
				Text("Invite"),
				hx.Post("/invite"),
				hx.Swap("outerHTML"),
				hx.Vals(string(must(json.Marshal(map[string]string{
					"csrf": csrf,
				})))),
			),
		),
	)
}

// InviteLinkPartial renders an invite link.
func InviteLinkPartial(token uuid.UUID) Node {
	u := fmt.Sprintf("/register/%s", token.String())

	return Div(
		P(Text("Copy and share this one-time link:")),
		A(ID("invite-link"),
			Class("hover:underline text-amber-600 font-semibold"),
			Href(u),
			Target("_blank"),
		),
		Script(Raw(fmt.Sprintf(`document.getElementById('invite-link').innerHTML = window.location.protocol + '//' + window.location.host + '%s';`, u))),
	)
}

// RegisterMain renders the registration page main content.
func RegisterMain(form contract.RegisterForm, errs url.Values, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"),
				Method("POST"),

				Label(Class("block w-full pt-2"), For("username"), Text("Username")),
				input("username", "text", "Username", form.Username, errs.Get("username"), true, false),

				Label(Class("block w-full pt-2"), For("password1"), Text("Password")),
				input("password1", "password", "Password", form.Password1, errs.Get("password1"), true, false),

				Label(Class("block w-full pt-2"), For("password2"), Text("Password again")),
				input("password2", "password", "Password again", form.Password2, errs.Get("password2"), true, false),

				Input(Type("hidden"), Name("csrf"), Value(csrf)),

				submitButton("Register"),
			),
		),
	)
}
