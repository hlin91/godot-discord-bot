package module

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Module struct {
	commands []*discordgo.ApplicationCommand
	handlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
}

func (m *Module) Load(session *discordgo.Session, commandHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate), register bool) {
	loadHandlers(commandHandlers, m.handlers)
	if register {
		registerCommands(session, m.commands)
	}
}

func CreateModule(commands []*discordgo.ApplicationCommand, handlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) Module {
	return Module{
		commands: commands,
		handlers: handlers,
	}
}

func loadHandlers(commandHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate), handlersToAdd map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) {
	for key, val := range handlersToAdd {
		if _, ok := commandHandlers[key]; ok {
			log.Panicf("handler for command '%v' is already registered", key)
		}
		commandHandlers[key] = val
	}
}

func registerCommands(session *discordgo.Session, commands []*discordgo.ApplicationCommand) {
	ug, err := session.UserGuilds(100, "", "")
	if err != nil {
		log.Panicf("could not retrieve user guilds: %v", err)
	}
	for _, v := range commands {
		for _, g := range ug {
			_, err := session.ApplicationCommandCreate(session.State.User.ID, g.ID, v)
			if err != nil {
				log.Printf("could not create '%v' command: %v", v.Name, err)
			}
		}
	}
}
