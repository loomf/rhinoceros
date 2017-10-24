package main

import (
	"fmt"
	"flag"
	"github.com/bwmarrin/discordgo"
	"os"
	"time"
)

var startTime time.Time

var session *discordgo.Session
var CHANNEL_ID string

var hosts map[string]string

func init() {
	startTime = time.Now()
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getChannelId(sess *discordgo.Session, guildname string, channelname string) (string, error) {
	var channelid string
	var guildid string
	guilds, err := sess.UserGuilds(100, "", "")
	panicOnErr(err)
	if err != nil {
		return "", err
	}
	for _, guild := range guilds {
		panicOnErr(err)
		if err != nil {
			return "", err
		}
		if guild.Name == guildname {
			guildid = guild.ID
			if err != nil {
				panicOnErr(err)
				return "", err
			}
		}
	}

	channels, err := sess.GuildChannels(guildid)
	if err != nil {
		panicOnErr(err)
		return "",err
	}
	for _, channel := range channels {
		channel, err = sess.Channel(channel.ID)
		panicOnErr(err)
		if err != nil {
			return "",err
		}
		if channel.Name == channelname {
			if err != nil {
				panicOnErr(err)
				return "",err
			}
			if channel.Type == discordgo.ChannelTypeGuildText {
				channelid = channel.ID
				return channelid, nil
			}
		}
	}

	return channelid, nil
}

func sendMessage(sess *discordgo.Session, message string, channelid string) {
	fmt.Println("SENDING MESSAGE:", message)
	_, err := sess.ChannelMessageSend(channelid, message)
	panicOnErr(err)
}

func handleMessages(sess *discordgo.Session, evt *discordgo.MessageCreate) {
	message := evt.Message
	messageChannel,_ := sess.Channel(message.ChannelID)
	if( messageChannel.Type == discordgo.ChannelTypeDM) {
		sendMessage(sess, message.Content, CHANNEL_ID)
	}
}

func main() {
	var (
		Token   = flag.String("t", "", "Discord Authentication Token")
		Guild   = flag.String("g", "", "Guild Name")
		Channel = flag.String("c", "", "Channel Name")
	)
	flag.Parse()
	if(*Token == "") {
		*Token = os.Getenv("TOKEN")
	}
	if(*Guild == "") {
		*Guild = os.Getenv("GUILD")
	}
	if(*Channel == "") {
		*Channel = os.Getenv("CHANNEL")
	}

	fmt.Println("Logging in...")
	var err error
	session, err = discordgo.New("Bot " + *Token)
	session.ShouldReconnectOnError = true

	session.AddHandler(handleMessages)

	panicOnErr(err)
	fmt.Println("Opening session...")
	err = session.Open()
	panicOnErr(err)

	CHANNEL_ID,err = getChannelId(session, *Guild, *Channel)
	panicOnErr(err)
	fmt.Println("Sleeping...")
	select {}
}
