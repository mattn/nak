package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mailru/easyjson"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nson"
	"github.com/urfave/cli/v2"
)

const CATEGORY_EVENT_FIELDS = "EVENT FIELDS"

var event = &cli.Command{
	Name:  "event",
	Usage: "generates an encoded event and either prints it or sends it to a set of relays",
	Description: `example usage (for sending directly to a relay with 'nostcat'):
		nak event -k 1 -c hello --envelope | nostcat wss://nos.lol
standalone:
		nak event -k 1 -c hello wss://nos.lol`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "sec",
			Usage:       "secret key to sign the event",
			DefaultText: "the key '1'",
			Value:       "0000000000000000000000000000000000000000000000000000000000000001",
		},
		&cli.BoolFlag{
			Name:  "envelope",
			Usage: "print the event enveloped in a [\"EVENT\", ...] message ready to be sent to a relay",
		},
		&cli.BoolFlag{
			Name:  "nson",
			Usage: "encode the event using NSON",
		},
		&cli.IntFlag{
			Name:        "kind",
			Aliases:     []string{"k"},
			Usage:       "event kind",
			DefaultText: "1",
			Value:       1,
			Category:    CATEGORY_EVENT_FIELDS,
		},
		&cli.StringFlag{
			Name:        "content",
			Aliases:     []string{"c"},
			Usage:       "event content",
			DefaultText: "hello from the nostr army knife",
			Value:       "hello from the nostr army knife",
			Category:    CATEGORY_EVENT_FIELDS,
		},
		&cli.StringSliceFlag{
			Name:     "tag",
			Aliases:  []string{"t"},
			Usage:    "sets a tag field on the event, takes a value like -t e=<id>",
			Category: CATEGORY_EVENT_FIELDS,
		},
		&cli.StringSliceFlag{
			Name:     "e",
			Usage:    "shortcut for --tag e=<value>",
			Category: CATEGORY_EVENT_FIELDS,
		},
		&cli.StringSliceFlag{
			Name:     "p",
			Usage:    "shortcut for --tag p=<value>",
			Category: CATEGORY_EVENT_FIELDS,
		},
		&cli.StringFlag{
			Name:        "created-at",
			Aliases:     []string{"time", "ts"},
			Usage:       "unix timestamp value for the created_at field",
			DefaultText: "now",
			Value:       "now",
			Category:    CATEGORY_EVENT_FIELDS,
		},
	},
	ArgsUsage: "[relay...]",
	Action: func(c *cli.Context) error {
		evt := nostr.Event{
			Kind:    c.Int("kind"),
			Content: c.String("content"),
			Tags:    make(nostr.Tags, 0, 3),
		}

		tags := make(nostr.Tags, 0, 5)
		for _, tagFlag := range c.StringSlice("tag") {
			// tags are in the format key=value
			spl := strings.Split(tagFlag, "=")
			if len(spl) == 2 && len(spl[0]) > 0 {
				tag := nostr.Tag{spl[0]}
				// tags may also contain extra elements separated with a ";"
				spl2 := strings.Split(spl[1], ";")
				tag = append(tag, spl2...)
				// ~
				tags = append(tags, tag)
			}
		}
		for _, etag := range c.StringSlice("e") {
			tags = append(tags, []string{"e", etag})
		}
		for _, ptag := range c.StringSlice("p") {
			tags = append(tags, []string{"p", ptag})
		}
		if len(tags) > 0 {
			for _, tag := range tags {
				evt.Tags = append(evt.Tags, tag)
			}
		}

		createdAt := c.String("created-at")
		ts := time.Now()
		if createdAt != "now" {
			if v, err := strconv.ParseInt(createdAt, 10, 64); err != nil {
				return fmt.Errorf("failed to parse timestamp '%s': %w", createdAt, err)
			} else {
				ts = time.Unix(v, 0)
			}
		}
		evt.CreatedAt = nostr.Timestamp(ts.Unix())

		if err := evt.Sign(c.String("sec")); err != nil {
			return fmt.Errorf("error signing with provided key: %w", err)
		}

		relays := c.Args().Slice()
		if len(relays) > 0 {
			fmt.Println(evt.String())
			for _, url := range relays {
				fmt.Fprintf(os.Stderr, "publishing to %s... ", url)
				if relay, err := nostr.RelayConnect(c.Context, url); err != nil {
					fmt.Fprintf(os.Stderr, "failed to connect: %s\n", err)
				} else {
					ctx, cancel := context.WithTimeout(c.Context, 10*time.Second)
					defer cancel()
					if status, err := relay.Publish(ctx, evt); err != nil {
						fmt.Fprintf(os.Stderr, "failed: %s\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "%s.\n", status)
					}
				}
			}
		} else {
			var result string
			if c.Bool("envelope") {
				j, _ := json.Marshal([]any{"EVENT", evt})
				result = string(j)
			} else if c.Bool("nson") {
				result, _ = nson.Marshal(&evt)
			} else {
				j, _ := easyjson.Marshal(&evt)
				result = string(j)
			}
			fmt.Println(result)
		}

		return nil
	},
}
