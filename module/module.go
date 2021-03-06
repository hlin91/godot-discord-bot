package module

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// Module groups together commands and handlers for registration with the discord session
type Module struct {
	commands          []*discordgo.ApplicationCommand
	handlers          map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
	componentHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
	modalHandlers     map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// Load registers the commands and handlers with the discord session
func (m *Module) Load(session *discordgo.Session, commandHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate), componentHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate), modalHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate), register bool) {
	loadHandlers(commandHandlers, m.handlers)
	loadHandlers(componentHandlers, m.componentHandlers)
	loadHandlers(modalHandlers, m.modalHandlers)
	if register {
		registerCommands(session, m.commands)
	}
}

// CreateModule constructs a Module with a given set of commands and handlers
func CreateModule(commands []*discordgo.ApplicationCommand, handlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate), componentHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate), modalHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)) Module {
	return Module{
		commands:          commands,
		handlers:          handlers,
		componentHandlers: componentHandlers,
		modalHandlers:     modalHandlers,
	}
}

// loadHandlers loads the command handlers into the discord session
func loadHandlers(commandHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate), handlersToAdd map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) {
	for key, val := range handlersToAdd {
		if _, ok := commandHandlers[key]; ok {
			log.Panicf("handler for command '%v' is already registered", key)
		}
		commandHandlers[key] = val
	}
}

// registerCommands registers the commands with the discord session
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
