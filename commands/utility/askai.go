package utility

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/arithefirst/whisker/helpers"
	"github.com/bwmarrin/discordgo"
)

var DefineMagicalAI = &discordgo.ApplicationCommand{
	Name:        "magical",
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

	if query == "" {
		helpers.IntRespondEph(s, i, "You must provide a query.")
		return
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

	// Safe username extraction
	var username string
	if i.User != nil {
		username = i.User.Username
	} else if i.Member != nil && i.Member.User != nil {
		username = i.Member.User.Username
	} else {
		username = "unknown"
	}

	// Build prompt
	wantLong := helpers.DecideLength(query)

	// Convert history to simple strings for context
	histText := make([]string, len(history))
	for idx, h := range history {
		histText[idx] = fmt.Sprintf("%s: %s", h.Author, h.Content)
	}

	prompt := helpers.FormatPrompt(query, username, histText, wantLong)

	// Ask AI
	ctx := context.Background()
	resp, err := helpers.GenerateContent(ctx, prompt)
	if err != nil {
		helpers.IntRespondEph(s, i, fmt.Sprintf("Error generating response: %v", err))
		log.Print(err)
		return
	}

	// Discord length cap
	if len(resp) > 1800 {
		resp = resp[:1790] + "…"
	}
	if resp == "" {
		helpers.IntRespondEph(s, i, "No response was generated.")
		return
	}

	err = helpers.IntRespond(s, i, resp)
	if err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func getRecentMessages(s *discordgo.Session, ic *discordgo.InteractionCreate, n int) ([]Message, error) {
	msgs, err := s.ChannelMessages(ic.ChannelID, n, "", "", "")
	if err != nil {
		return nil, err
	}

	history := make([]Message, len(msgs))
	for i, m := range msgs {
		if m.Author == nil {
			continue
		}
		history[i] = Message{
			Author:   m.Author.Username,
			AuthorID: m.Author.ID,
			Content:  m.Content,
		}
	}

	slices.Reverse(history) // oldest → newest
	return history, nil
}
