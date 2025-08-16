package utility

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/arithefirst/whisker/helpers"
	"github.com/bwmarrin/discordgo"
)

// Slash command definition
var DefineMagicalAI = &discordgo.ApplicationCommand{
	Name:        "magicalcat",
	Description: "Ask the Magical Cat anything!",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "query",
			Description: "Your question",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        "limit",
			Description: "Number of recent messages to include as context",
			Type:        discordgo.ApplicationCommandOptionInteger,
		},
	},
}

// Minimal message struct
type Message struct {
	Author   string
	AuthorID string
	Content  string
}

func MagicalAI(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var query string
	var limit int

	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "query":
			query = opt.StringValue()
		case "limit":
			limit = int(opt.IntValue())
		}
	}

	if limit <= 0 {
		limit = 5
	}

	history, err := getRecentMessages(s, i, limit)
	if err != nil {
		helpers.IntRespondEph(s, i, fmt.Sprintf("Error generating response: %v", err))
		log.Printf("getRecentMessages: %v", err)
		return
	}

	prompt := fmt.Sprintf(
		"history = %v\n\nYou're given the last %v messages from a Discord channel. Your job is to answer the query \"%v\" by \"%v\" within 1800 words",
		history, limit, query, i.User.Username,
	)

	log.Println(prompt)
	ctx := context.Background()
	resp, err := helpers.GenerateContent(ctx, prompt)
	if err != nil {
		helpers.IntRespondEph(s, i, fmt.Sprintf("Error generating response: %v", err))
		log.Print(err)
		return
	}
	helpers.IntRespond(s, i, resp)
}

// Fetch N most recent messages
func getRecentMessages(s *discordgo.Session, ic *discordgo.InteractionCreate, n int) ([]Message, error) {
	msgs, err := s.ChannelMessages(ic.ChannelID, n, "", "", "")
	if err != nil {
		return nil, err
	}

	history := make([]Message, len(msgs))
	for i, m := range msgs {
		history[i] = Message{
			Author:   m.Author.Username,
			AuthorID: m.Author.ID,
			Content:  m.Content,
		}
	}

	slices.Reverse(history) // oldest → newest
	return history, nil
}
