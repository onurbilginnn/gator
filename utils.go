package main

import (
	"context"
	"strconv"

	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/onurbilginnn/gator/internal/database"

	"github.com/onurbilginnn/internal/config"
)

type state struct {
	Config *config.Config
	db     *database.Queries
}

type command struct {
	Name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (commands *commands) run(state *state, cmd command) error {
	handler, exists := commands.commands[cmd.Name]
	if !exists {
		fmt.Printf("unknown command: %s\n", cmd.Name)
		os.Exit(1)
	}
	return handler(state, cmd)
}

func (commands *commands) register(name string, handler func(*state, command) error) {
	if commands.commands == nil {
		commands.commands = make(map[string]func(*state, command) error)
	}
	commands.commands[name] = handler
}

func handlerLogin(state *state, command command) error {
	username := getArgFromCmd(command, "username")
	existedUser, getErr := state.db.GetUserByName(context.Background(), username)
	if getErr != nil {
		fmt.Printf("User not exists can not login: %s\n", username)
		os.Exit(1)
	}
	err := state.Config.SetUser(existedUser.Name)
	if err != nil {
		fmt.Printf("failed to set user: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("The user has been set: %s\n", existedUser.Name)
	return nil
}

func handlerRegister(state *state, command command) error {
	username := getArgFromCmd(command, "username")
	existedUser, getErr := state.db.GetUserByName(context.Background(), username)
	if getErr == nil {
		fmt.Printf("User already exists: %s\n", existedUser.Name)
		os.Exit(1)
	}
	user := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	}
	_, err := state.db.CreateUser(context.Background(), user)
	if err != nil {
		fmt.Printf("failed to register user: %v\n", err)
		os.Exit(1)
	}
	state.Config.SetUser(username)
	fmt.Printf("The user has been registered: %s\n", user.Name)
	fmt.Printf("ID: %s\n", user.ID)
	fmt.Printf("Create Date: %s\n", user.CreatedAt)
	fmt.Printf("Update Date: %s\n", user.UpdatedAt)
	return nil
}

func handlerReset(state *state, _ command) error {
	err := state.db.ResetUsers(context.Background())
	if err != nil {
		fmt.Printf("failed to reset users: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All users have been reset.")
	return nil
}

func handlerGetUsers(state *state, _ command) error {
	users, err := state.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("failed to get users: %v\n", err)
		os.Exit(1)
	}
	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}
	currentUser := state.Config.CurrentUsername
	for _, user := range users {
		userText := user.Name
		if userText == currentUser {
			userText += " (current)"
		}
		fmt.Printf("* %s\n", userText)
	}
	return nil
}

func handlerAggregate(state *state, command command) error {
	time_between_reqs := getArgFromCmd(command, "time_between_reqs")
	timeDuration, timeConvertErr := time.ParseDuration(time_between_reqs)
	if timeConvertErr != nil {
		fmt.Printf("failed to parse time duration: %v\n", timeConvertErr)
		os.Exit(1)
	}
	fmt.Printf("Collecting feeds every %s\n", timeDuration)
	ticker := time.NewTicker(timeDuration)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		scrapeFeeds(state)
	}
}

func handlerAddFeed(state *state, command command, currentUser database.User) error {
	args := getArgsFromCmd(command, 2)
	feedName := args[0]
	feedUrl := args[1]
	existedFeed, getErr := state.db.GetFeedByUrl(context.Background(), feedUrl)
	if getErr == nil {
		fmt.Printf("Feed already exists: %s\n", existedFeed.Name)
		os.Exit(1)
	}
	feed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    currentUser.ID,
	}
	_, err := state.db.CreateFeed(context.Background(), feed)
	if err != nil {
		fmt.Printf("failed to add feed: %v\n", err)
		os.Exit(1)
	}
	feedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, followErr := state.db.CreateFeedFollow(context.Background(), feedFollow)
	if followErr != nil {
		fmt.Printf("failed to follow feed: %v\n", followErr)
		os.Exit(1)
	}
	return nil
}

func handlerGetFeeds(state *state, _ command) error {
	feeds, err := state.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Printf("failed to get feeds: %v\n", err)
		os.Exit(1)
	}
	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}
	for _, feed := range feeds {
		user, err := state.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			fmt.Printf("failed to get user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("* %s (URL: %s, User: %s)\n", feed.Name, feed.Url, user.Name)
	}
	return nil
}

func handlerFollowFeed(state *state, command command, currentUser database.User) error {
	args := getArgsFromCmd(command, 1)
	feedUrl := args[0]
	currentFeed, feedErr := state.db.GetFeedByUrl(context.Background(), feedUrl)
	if feedErr != nil {
		fmt.Printf("Feed not found: %s\n", feedUrl)
		os.Exit(1)
	}
	feedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    currentUser.ID,
		FeedID:    currentFeed.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := state.db.CreateFeedFollow(context.Background(), feedFollow)
	if err != nil {
		fmt.Printf("failed to follow feed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s user successfully followed feed: %s\n", currentUser.Name, currentFeed.Name)
	return nil
}

func handlerGetFeedFollows(state *state, _ command, currentUser database.User) error {
	feedFollows, err := state.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		fmt.Printf("failed to get feed follows: %v\n", err)
		os.Exit(1)
	}
	if len(feedFollows) == 0 {
		fmt.Println("No feed follows found.")
		return nil
	}
	for _, feedFollow := range feedFollows {
		feed, err := state.db.GetFeedById(context.Background(), feedFollow.FeedID)
		if err != nil {
			fmt.Printf("failed to get feed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("* Feed Name: %s, Feed URL: %s\n", feed.Name, feed.Url)
	}
	return nil
}

func handlerUnfollowFeed(state *state, command command, currentUser database.User) error {
	feedUrl := getArgFromCmd(command, "feed_url")
	feed, feedErr := state.db.GetFeedByUrl(context.Background(), feedUrl)
	if feedErr != nil {
		fmt.Printf("Feed not found: %s\n", feedUrl)
		os.Exit(1)
	}

	unfollowErr := state.db.UnfollowFeed(context.Background(), database.UnfollowFeedParams{
		UserID: currentUser.ID,
		FeedID: feed.ID,
	})
	if unfollowErr != nil {
		fmt.Printf("Failed to unfollow feed: %v\n", unfollowErr)
		os.Exit(1)
	}

	fmt.Printf("%s user successfully unfollowed feed: %s\n", currentUser.Name, feed.Name)
	return nil
}

func handlerBrowsePosts(state *state, command command) error {
	limitStr := getArgFromCmdWithDefault(command, "2")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		fmt.Println("failed to convert limit to integer using default value 2")
		limit = 2
	}

	posts, err := state.db.GetPosts(context.Background(), int32(limit))
	if err != nil {
		fmt.Printf("failed to get posts: %v\n", err)
		os.Exit(1)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found.")
		return nil
	}

	for _, post := range posts {
		fmt.Printf("* Title: %s\n", post.Title)
		fmt.Printf("  URL: %s\n", post.Url)
		fmt.Printf("  Description: %s\n", post.Description.String)
		fmt.Printf("  Published At: %s\n", post.PublishedAt)
		fmt.Println()
	}

	return nil
}
