package main

func pingCommand(ctx Context) error {
	ctx.Reply("Pong !")
	return nil
}
