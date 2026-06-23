package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/onurbilginnn/gator/internal/database"
	"github.com/onurbilginnn/internal/config"
)

func main() {
	readConfig, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading %v\n", err)
		return
	}
	state := &state{Config: readConfig}
	dbUrl := state.Config.DBUrl
	if dbUrl == "" {
		fmt.Println("DB URL is not set in the config file")
		os.Exit(1)
	}
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Printf("Error connecting to the database: %v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	state.db = dbQueries
	commands := &commands{}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerGetUsers)
	commands.register("agg", handlerAggregate)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerGetFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollowFeed))
	commands.register("following", middlewareLoggedIn(handlerGetFeedFollows))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	commands.register("browse", handlerBrowsePosts)
	if len(os.Args) < 2 {
		fmt.Println("No command provided")
		os.Exit(1)
	}
	cmd := command{Name: os.Args[1], args: os.Args[2:]}
	err = commands.run(state, cmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
