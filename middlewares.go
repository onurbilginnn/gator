package main

import (
	"context"

	"github.com/onurbilginnn/gator/internal/database"
)

func middlewareLoggedIn(handler func(state *state, cmd command, user database.User) error) func(*state, command) error {
	return func(state *state, cmd command) error {
		currentUser, userErr := state.db.GetUserByName(context.Background(), state.Config.CurrentUsername)
		if userErr != nil {
			return userErr
		}
		return handler(state, cmd, currentUser)
	}
}
