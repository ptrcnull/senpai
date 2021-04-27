package senpai

import (
	"fmt"
	"strings"
	"time"

	"git.sr.ht/~taiite/senpai/irc"
	"git.sr.ht/~taiite/senpai/ui"
)

type command struct {
	AllowHome bool
	MinArgs   int
	MaxArgs   int
	Usage     string
	Desc      string
	Handle    func(app *App, buffer string, args []string) error
}

type commandSet map[string]*command

var commands commandSet

func init() {
	commands = commandSet{
		"": {
			MinArgs: 1,
			MaxArgs: 1,
			Handle:  commandDo,
		},
		"HELP": {
			AllowHome: true,
			MaxArgs:   1,
			Usage:     "[command]",
			Desc:      "show the list of commands, or how to use the given one",
			Handle:    commandDoHelp,
		},
		"JOIN": {
			AllowHome: true,
			MinArgs:   1,
			MaxArgs:   2,
			Usage:     "<channels> [keys]",
			Desc:      "join a channel",
			Handle:    commandDoJoin,
		},
		"ME": {
			AllowHome: true,
			MinArgs:   1,
			MaxArgs:   1,
			Usage:     "<message>",
			Desc:      "send an action (reply to last query if sent from home)",
			Handle:    commandDoMe,
		},
		"MSG": {
			AllowHome: true,
			MinArgs:   2,
			MaxArgs:   2,
			Usage:     "<target> <message>",
			Desc:      "send a message to the given target",
			Handle:    commandDoMsg,
		},
		"NAMES": {
			Desc:   "show the member list of the current channel",
			Handle: commandDoNames,
		},
		"PART": {
			AllowHome: true,
			MaxArgs:   2,
			Usage:     "[channel] [reason]",
			Desc:      "part a channel",
			Handle:    commandDoPart,
		},
		"QUIT": {
			AllowHome: true,
			MaxArgs:   1,
			Usage:     "[reason]",
			Desc:      "quit senpai",
			Handle:    commandDoQuit,
		},
		"QUOTE": {
			AllowHome: true,
			MinArgs:   1,
			MaxArgs:   1,
			Usage:     "<raw message>",
			Desc:      "send raw protocol data",
			Handle:    commandDoQuote,
		},
		"REPLY": {
			AllowHome: true,
			MinArgs:   1,
			MaxArgs:   1,
			Usage:     "<message>",
			Desc:      "reply to the last query",
			Handle:    commandDoR,
		},
		"TOPIC": {
			MaxArgs: 1,
			Usage:   "[topic]",
			Desc:    "show or set the topic of the current channel",
			Handle:  commandDoTopic,
		},
		"BUFFER": {
			AllowHome: true,
			MinArgs:   1,
			MaxArgs:   1,
			Usage:     "<name>",
			Desc:      "switch to the buffer containing a substring",
			Handle:    commandDoBuffer,
		},
	}
}

func commandDo(app *App, buffer string, args []string) (err error) {
	app.s.PrivMsg(buffer, args[0])
	if !app.s.HasCapability("echo-message") {
		buffer, line, _ := app.formatMessage(irc.MessageEvent{
			User:            &irc.Prefix{Name: app.s.Nick()},
			Target:          buffer,
			TargetIsChannel: true,
			Command:         "PRIVMSG",
			Content:         args[0],
			Time:            time.Now(),
		})
		app.win.AddLine(buffer, false, line)
	}
	return
}

func commandDoHelp(app *App, buffer string, args []string) (err error) {
	t := time.Now()
	if len(args) == 0 {
		app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
			At:   t,
			Head: "--",
			Body: "Available commands:",
		})
		for cmdName, cmd := range commands {
			if cmd.Desc == "" {
				continue
			}
			app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
				At:   t,
				Body: fmt.Sprintf("  \x02%s\x02 %s", cmdName, cmd.Usage),
			})
			app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
				At:   t,
				Body: fmt.Sprintf("    %s", cmd.Desc),
			})
			app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
				At: t,
			})
		}
	} else {
		search := strings.ToUpper(args[0])
		found := false
		app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
			At:   t,
			Head: "--",
			Body: fmt.Sprintf("Commands that match \"%s\":", search),
		})
		for cmdName, cmd := range commands {
			if !strings.Contains(cmdName, search) {
				continue
			}
			app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
				At:   t,
				Body: fmt.Sprintf("\x02%s\x02 %s", cmdName, cmd.Usage),
			})
			app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
				At:   t,
				Body: fmt.Sprintf("  %s", cmd.Desc),
			})
			app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
				At: t,
			})
			found = true
		}
		if !found {
			app.win.AddLine(app.win.CurrentBuffer(), false, ui.Line{
				At:   t,
				Body: fmt.Sprintf("  no command matches %q", args[0]),
			})
		}
	}
	return
}

func commandDoJoin(app *App, buffer string, args []string) (err error) {
	app.s.Join(args[0])
	return
}

func commandDoMe(app *App, buffer string, args []string) (err error) {
	if buffer == Home {
		buffer = app.lastQuery
	}
	content := fmt.Sprintf("\x01ACTION %s\x01", args[0])
	app.s.PrivMsg(buffer, content)
	if !app.s.HasCapability("echo-message") {
		buffer, line, _ := app.formatMessage(irc.MessageEvent{
			User:            &irc.Prefix{Name: app.s.Nick()},
			Target:          buffer,
			TargetIsChannel: true,
			Command:         "PRIVMSG",
			Content:         content,
			Time:            time.Now(),
		})
		app.win.AddLine(buffer, false, line)
	}
	return
}

func commandDoMsg(app *App, buffer string, args []string) (err error) {
	target := args[0]
	content := args[1]
	app.s.PrivMsg(target, content)
	if !app.s.HasCapability("echo-message") {
		buffer, line, _ := app.formatMessage(irc.MessageEvent{
			User:            &irc.Prefix{Name: app.s.Nick()},
			Target:          target,
			TargetIsChannel: true,
			Command:         "PRIVMSG",
			Content:         content,
			Time:            time.Now(),
		})
		app.win.AddLine(buffer, false, line)
	}
	return
}

func commandDoNames(app *App, buffer string, args []string) (err error) {
	var sb strings.Builder
	sb.WriteString("\x0314Names: ")
	for _, name := range app.s.Names(buffer) {
		if name.PowerLevel != "" {
			sb.WriteString("\x033")
			sb.WriteString(name.PowerLevel)
			sb.WriteString("\x0314")
		}
		sb.WriteString(name.Name.Name)
		sb.WriteRune(' ')
	}
	body := sb.String()
	app.win.AddLine(buffer, false, ui.Line{
		At:   time.Now(),
		Head: "--",
		Body: body[:len(body)-1],
	})
	return
}

func commandDoPart(app *App, buffer string, args []string) (err error) {
	channel := buffer
	reason := ""
	if 0 < len(args) {
		if app.s.IsChannel(args[0]) {
			channel = args[0]
			if 1 < len(args) {
				reason = args[1]
			}
		} else {
			reason = args[0]
		}
	}

	if channel != Home {
		app.s.Part(channel, reason)
	} else {
		err = fmt.Errorf("cannot part home!")
	}
	return
}

func commandDoQuit(app *App, buffer string, args []string) (err error) {
	reason := ""
	if 0 < len(args) {
		reason = args[0]
	}
	if app.s != nil {
		app.s.Quit(reason)
	}
	app.win.Exit()
	return
}

func commandDoQuote(app *App, buffer string, args []string) (err error) {
	app.s.SendRaw(args[0])
	return
}

func commandDoR(app *App, buffer string, args []string) (err error) {
	app.s.PrivMsg(app.lastQuery, args[0])
	if !app.s.HasCapability("echo-message") {
		buffer, line, _ := app.formatMessage(irc.MessageEvent{
			User:            &irc.Prefix{Name: app.s.Nick()},
			Target:          app.lastQuery,
			TargetIsChannel: true,
			Command:         "PRIVMSG",
			Content:         args[0],
			Time:            time.Now(),
		})
		app.win.AddLine(buffer, false, line)
	}
	return
}

func commandDoTopic(app *App, buffer string, args []string) (err error) {
	if len(args) == 0 {
		var body string

		topic, who, at := app.s.Topic(buffer)
		if who == nil {
			body = fmt.Sprintf("\x0314Topic: %s", topic)
		} else {
			body = fmt.Sprintf("\x0314Topic (by %s, %s): %s", who, at.Local().Format("Mon Jan 2 15:04:05"), topic)
		}
		app.win.AddLine(buffer, false, ui.Line{
			At:   time.Now(),
			Head: "--",
			Body: body,
		})
	} else {
		app.s.SetTopic(buffer, args[0])
	}
	return
}

func parseCommand(s string) (command, args string) {
	if s == "" {
		return
	}

	if s[0] != '/' {
		args = s
		return
	}

	i := strings.IndexByte(s, ' ')
	if i < 0 {
		i = len(s)
	}

	command = strings.ToUpper(s[1:i])
	args = strings.TrimLeft(s[i:], " ")

	return
}

func (app *App) handleInput(buffer, content string) error {
	cmdName, rawArgs := parseCommand(content)

	var chosenCMDName string
	var ok bool
	for key := range commands {
		if cmdName == "" && key != "" {
			continue
		}
		if !strings.HasPrefix(key, cmdName) {
			continue
		}
		if ok {
			return fmt.Errorf("ambiguous command %q (could mean %v or %v)", cmdName, chosenCMDName, key)
		}
		chosenCMDName = key
		ok = true
	}
	if !ok {
		return fmt.Errorf("command %q doesn't exist", cmdName)
	}

	cmd := commands[chosenCMDName]

	var args []string
	if rawArgs != "" && cmd.MaxArgs != 0 {
		args = strings.SplitN(rawArgs, " ", cmd.MaxArgs)
	}

	if len(args) < cmd.MinArgs {
		return fmt.Errorf("usage: %s %s", cmdName, cmd.Usage)
	}
	if buffer == Home && !cmd.AllowHome {
		return fmt.Errorf("command %q cannot be executed from home", cmdName)
	}

	return cmd.Handle(app, buffer, args)
}

func commandDoBuffer(app *App, buffer string, args []string) error {
	name := args[0]
	if !app.win.JumpBuffer(args[0]) {
		return fmt.Errorf("none of the buffers match %q", name)
	}

	return nil
}
