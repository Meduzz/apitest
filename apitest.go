package main

import "meduzz.github.com/apitest/commands"

func main() {
	err := commands.Root.Execute()

	if err != nil {
		panic(err)
	}
}
