package utility

import (
	"bytes"
	"fmt"
	"image"
	"net/http"
	"strings"

	"github.com/arithefirst/whisker/helpers"
	colors "github.com/arithefirst/whisker/helpers/embedColors"
	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/imaging"
)

var DefineImageManipulation = &discordgo.ApplicationCommand{
	Name:        "image",
	Description: "Edit image with funcs ~ sharpen, filters, resize etc.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "src",
			Description: "provide a image (optional)",
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Required:    false,
		},
		{
			Name:        "action",
			Description: "supports actions like sharpen, crop, resize etc.",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Sharpen",
					Value: "sharpen",
				},
				{
					Name:  "Gamma",
					Value: "gamma",
				},
				{
					Name:  "Brightness",
					Value: "brightness",
				},
				{
					Name:  "Saturation",
					Value: "saturation",
				},
				{
					Name:  "Blur",
					Value: "blur",
				},
				{
					Name:  "Contrast",
					Value: "contrast",
				},
			},
		},
		{
			Name:        "sigma",
			Description: "",
			Type:        discordgo.ApplicationCommandOptionNumber,
			Required:    false,
		},
	},
}

func ImageManipulation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, option := range options {
		optionMap[option.Name] = option
	}

	var srcImage image.Image
	var dstImage image.Image
	var action string
	var sigma float64
	var err error

	if opt, ok := optionMap["src"]; ok {
		// Attachment ID
		attachmentID := opt.Value.(string)
		// MessageAttachment Object (Contains it's URL)
		attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentID]
		if !strings.HasPrefix(attachment.ContentType, "image/") {
			helpers.IntRespond(s, i, "Make sure it's an image : )")
		}
		srcImage, err = DownloadImage(attachment)
		if err != nil {
			helpers.IntRespondEph(s, i, err.Error())
			return
		}
	}

	if opt, ok := optionMap["sigma"]; ok {
		sigma = opt.FloatValue()
	}

	if opt, ok := optionMap["action"]; ok {
		action = opt.StringValue()
	}

	dstImage = SigmaImageHelper(srcImage, action, sigma)

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, dstImage, imaging.JPEG)

	filename := "edited.png"
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s - %f", action, sigma),
		Image: &discordgo.MessageEmbedImage{
			URL: "attachment://" + filename,
		},
		Color: colors.Primary,
	}

	_, err = s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Files: []*discordgo.File{
			{
				Name:   filename,
				Reader: bytes.NewReader(buf.Bytes()),
			},
		},
	})
}

func DownloadImage(attachmentObj *discordgo.MessageAttachment) (image.Image, error) {
	resp, err := http.Get(attachmentObj.URL)
	if err != nil {
		err := fmt.Errorf("error downloading image: %w", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("download failed with an error: %d", resp.StatusCode)
		return nil, err
	}
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		err := fmt.Errorf("error decoding image: %w", err)
		return nil, err
	}
	return img, nil
}

func SigmaImageHelper(srcImage image.Image, action string, sigma float64) image.Image {
	var dstImage image.Image
	switch action {
	case "sharpen":
		dstImage = imaging.Sharpen(srcImage, sigma)
	case "gamma":
		dstImage = imaging.AdjustGamma(srcImage, sigma)
	case "contrast":
		dstImage = imaging.AdjustContrast(srcImage, sigma)
	case "brightness":
		dstImage = imaging.AdjustBrightness(srcImage, sigma)
	case "saturation":
		dstImage = imaging.AdjustSaturation(srcImage, sigma)
	case "blur":
		dstImage = imaging.Blur(srcImage, sigma)
	}
	return dstImage
}
