package main

import (
	"errors"
	"flag"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"strings"
	"time"
)

var logger *log.Logger
var startTime time.Time

var session *discordgo.Session
var CHANNEL_ID string

var hosts map[string]string

func init() {
	logger = log.New(os.Stderr, "  ", log.Ldate|log.Ltime)
	startTime = time.Now()
}

func logDebug(v ...interface{}) {
	logger.SetPrefix("DEBUG ")
	logger.Println(v...)
}

func logInfo(v ...interface{}) {
	logger.SetPrefix("INFO  ")
	logger.Println(v...)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

/* Tries to call a method and checking if the method returned an error, if it
did check to see if it's HTTP 502 from the Discord API and retry for
`attempts` number of times. */
func retryOnBadGateway(f func() error) {
	var err error
	for i := 0; i < 3; i++ {
		err = f()
		if err != nil {
			if strings.HasPrefix(err.Error(), "HTTP 502") {
				// If the error is Bad Gateway, try again after 1 sec.
				time.Sleep(1 * time.Second)
				continue
			} else {
				// Otherwise panic !
				panicOnErr(err)
			}
		} else {
			// In case of no error, return.
			return
		}
	}
}

func getChannelId(sess *discordgo.Session, guildname string, channelname string) string {
	var channelid string
	retryOnBadGateway(func() error {
		var guildid string
		guilds, err := sess.UserGuilds()
		if err != nil {
			return err
		}
		for _, guild := range guilds {
			guild, err = sess.Guild(guild.ID)
			if err != nil {
				return err
			}
			if guild.Name == guildname {
				guildid = guild.ID
				if err != nil {
					return err
				}
			}
		}

		channels, err := sess.GuildChannels(guildid)
		if err != nil {
			return err
		}
		for _, channel := range channels {
			channel, err = sess.Channel(channel.ID)
			if err != nil {
				return err
			}
			if channel.Name == channelname {
				if err != nil {
					return err
				}
				if channel.Type == "text" {
					channelid = channel.ID
					return nil
				}
			}
		}
		return errors.New("Channel " + channelname + " not found in guild " + guildname)
	})

	return channelid
}

func fetchPrimaryTextChannelID(sess *discordgo.Session) string {
	var channelid string
	retryOnBadGateway(func() error {
		guilds, err := sess.UserGuilds()
		if err != nil {
			return err
		}
		guild, err := sess.Guild(guilds[0].ID)
		if err != nil {
			return err
		}
		channels, err := sess.GuildChannels(guild.ID)
		if err != nil {
			return err
		}
		for _, channel := range channels {
			channel, err = sess.Channel(channel.ID)
			if err != nil {
				return err
			}
			if channel.Type == "text" {
				channelid = channel.ID
				return nil
			}
		}
		return errors.New("No primary channel found")
	})
	return channelid
}

func sendMessage(sess *discordgo.Session, message string, channelid string) {
	logInfo("SENDING MESSAGE:", message)
	retryOnBadGateway(func() error {
		_, err := sess.ChannelMessageSend(channelid, message)
		return err
	})
}

func handleMessages(sess *discordgo.Session, evt *discordgo.MessageCreate) {
	message := evt.Message
	messageChannel,_ := sess.Channel(message.ChannelID)
	if( messageChannel.IsPrivate ) {
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

	logInfo("Logging in...")
	var err error
	session, err = discordgo.New(*Token)
	session.ShouldReconnectOnError = true

	session.AddHandler(handleMessages)

	panicOnErr(err)
	logInfo("Opening session...")
	err = session.Open()
	panicOnErr(err)

	CHANNEL_ID = getChannelId(session, *Guild, *Channel)
	logInfo("Sleeping...")
	select {}
}
