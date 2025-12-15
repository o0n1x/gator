package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/o0n1x/aggreGator/internal/cli"
	"github.com/o0n1x/aggreGator/internal/config"
	"github.com/o0n1x/aggreGator/internal/database"
)

func main() {

	//config and db connection
	cnfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cnfg.DB_URL)
	if err != nil {
		fmt.Printf("DB Error: %v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	state := cli.State{
		State: &cnfg,
		DB:    dbQueries,
	}

	//commands
	commands := cli.Commands{
		Commands: make(map[string]func(*cli.State, cli.Command) error),
	}
	commands.Register("login", cli.HandlerLogin)
	commands.Register("register", cli.HandlerRegister)
	commands.Register("reset", cli.HandlerReset)
	commands.Register("users", cli.HandlerUsers)
	commands.Register("agg", cli.HandlerAgg)
	commands.Register("addfeed", cli.HandlerAddFeed)

	//command executing
	if len(os.Args) < 2 {
		fmt.Printf("Error: %v\n", "Invalid input No arguments")
		os.Exit(1)
	}

	err = commands.Run(&state, cli.Command{Name: os.Args[1], Args: os.Args[2:]})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

}
