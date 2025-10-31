package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/msoap/raphanus"
	// tgbotapi "gopkg.in/telegram-bot-api.v2" // inline keyboards not supported in v2
)

// Ctx - context for bot command function (users, command, args, ...)
type Ctx struct {
	appConfig      *Config           // configuration
	users          *Users            // all users
	commands       Commands          // all chat commands
	userID         int               // current user
	allowExec      bool              // is user authorized
	messageCmd     string            // command name
	messageArgs    string            // command arguments
	messageSignal  chan<- BotMessage // for send telegram messages
	chatID         int               // chat for send replay
	exitSignal     chan<- struct{}   // for signal for terminate bot
	cache          *raphanus.DB      // cache for commands output
	cacheTTL       int               // cache timeout
	oneThreadMutex *sync.Mutex       // mutex for run shell commands in one thread
	uploadedFile   string            // path to uploaded file (if any)
	database       *Database         // database for logging
}

// /auth and /authroot - authorize users
func cmdAuth(ctx Ctx) (replayMsg string) {
	forRoot := ctx.messageCmd == "/authroot"

	if ctx.messageArgs == "" {

		replayMsg = "See code in terminal with shell2telegram or ask code from root user and type:\n" + ctx.messageCmd + " code"
		authCode := ctx.users.DoLogin(ctx.userID, forRoot)

		rootRoleStr := ""
		if forRoot {
			rootRoleStr = "root "
		}
		secretCodeMsg := fmt.Sprintf("Request %saccess for %s. Code: %s\n", rootRoleStr, ctx.users.String(ctx.userID), authCode)
		fmt.Print(secretCodeMsg)
		ctx.users.BroadcastForRoots(ctx.messageSignal, secretCodeMsg, 0)

	} else {
		if ctx.users.IsValidCode(ctx.userID, ctx.messageArgs, forRoot) {
			ctx.users.SetAuthorized(ctx.userID, forRoot)
			if forRoot {
				replayMsg = fmt.Sprintf("You (%s) authorized as root.", ctx.users.String(ctx.userID))
				log.Print("root authorized: ", ctx.users.String(ctx.userID))
			} else {
				replayMsg = fmt.Sprintf("You (%s) authorized.", ctx.users.String(ctx.userID))
				log.Print("authorized: ", ctx.users.String(ctx.userID))
			}
		} else {
			replayMsg = "Code is not valid."
		}
	}

	return replayMsg
}

// /help
func cmdHelp(ctx Ctx) (replayMsg string) {
	helpMsg := []string{}

	if ctx.allowExec {
		for cmd, shellCmdRow := range ctx.commands {
			description := shellCmdRow.description
			if description == "" {
				description = shellCmdRow.shellCmd
			}
			helpMsg = append(helpMsg, cmd+" → "+description)
		}
	}
	sort.Strings(helpMsg)

	if !ctx.appConfig.isPublicBot {
		helpMsg = append(helpMsg,
			"/auth [code] → authorize user",
			"/authroot [code] → authorize user as root",
		)
	}

	if ctx.allowExec {
		helpMsg = append(helpMsg,
			"/history → show recent command history",
			"/history export → export full history to file",
		)
	}

	if ctx.users.IsRoot(ctx.userID) {
		helpMsgForRoot := []string{
			"/shell2telegram ban <user_id|username> → ban user",
			"/shell2telegram broadcast_to_root <message> → send message to all root users in private chat",
			"/shell2telegram desc <bot description> → set bot description",
			"/shell2telegram message_to_user <user_id|username> <message> → send message to user in private chat",
			"/shell2telegram rm </command> → delete command",
			"/shell2telegram search <query> → search users by name/id",
			"/shell2telegram stat → get stat about users",
			"/shell2telegram version → show version",
		}
		if ctx.appConfig.addExit {
			helpMsgForRoot = append(helpMsgForRoot, "/shell2telegram exit → terminate bot")
		}
		sort.Strings(helpMsgForRoot)

		helpMsg = append(helpMsg, helpMsgForRoot...)
	}

	if ctx.appConfig.description != "" {
		replayMsg = ctx.appConfig.description
	} else {
		replayMsg = "This bot created with shell2telegram"
	}
	replayMsg += "\n\n" +
		"available commands:\n" +
		strings.Join(helpMsg, "\n")

	return replayMsg
}

// all commands from command-line
func cmdUser(ctx Ctx) {
	if cmd, found := ctx.commands[ctx.messageCmd]; found {
		go func() {
			if ctx.appConfig.oneThread {
				ctx.oneThreadMutex.Lock()
			}
			replayMsgRaw, executionTime, success := execShellWithFile(
				cmd.shellCmd,
				ctx.messageArgs,
				ctx.commands[ctx.messageCmd].vars,
				ctx.userID,
				ctx.chatID,
				ctx.users.list[ctx.userID].UserName,
				ctx.users.list[ctx.userID].FirstName+" "+ctx.users.list[ctx.userID].LastName,
				ctx.cache,
				ctx.cacheTTL,
				ctx.appConfig,
				ctx.uploadedFile,
			)
			if ctx.appConfig.oneThread {
				ctx.oneThreadMutex.Unlock()
			}

			// Log to database if enabled
			if ctx.database != nil {
				ctx.database.LogCommand(
					ctx.userID,
					ctx.users.list[ctx.userID].UserName,
					ctx.messageCmd,
					ctx.messageArgs,
					replayMsgRaw,
					executionTime,
					success,
				)
			}

			// Note: Inline keyboard feature requires telegram-bot-api v5+ (future enhancement)

			sendMessage(ctx.messageSignal, ctx.chatID, replayMsgRaw, cmd.isMarkdown, cmd.isFile)
		}()
	}
}

// /shell2telegram stat
func cmdShell2telegramStat(ctx Ctx) (replayMsg string) {
	for userID := range ctx.users.list {
		replayMsg += ctx.users.StringVerbose(userID) + "\n"
	}

	return replayMsg
}

// /shell2telegram search
func cmdShell2telegramSearch(ctx Ctx) (replayMsg string) {
	query := ctx.messageArgs

	if query == "" {
		return "Please set query: /shell2telegram search <query>"
	}

	for _, userID := range ctx.users.Search(query) {
		replayMsg += ctx.users.StringVerbose(userID) + "\n"
	}

	return replayMsg
}

// /shell2telegram ban
func cmdShell2telegramBan(ctx Ctx) (replayMsg string) {
	userName := ctx.messageArgs

	if userName == "" {
		return "Please set user_id or login: /shell2telegram ban <user_id|username>"
	}

	userID := ctx.users.FindByIDOrUserName(userName)

	if userID > 0 && ctx.users.BanUser(userID) {
		replayMsg = fmt.Sprintf("User %s banned", ctx.users.String(userID))
	} else {
		replayMsg = "User not found"
	}

	return replayMsg
}

// set bot description
func cmdShell2telegramDesc(ctx Ctx) (replayMsg string) {
	description := ctx.messageArgs

	if description == "" {
		return "Please set description: /shell2telegram desc <bot description>"
	}

	ctx.appConfig.description = description
	replayMsg = "Bot description set to: " + description

	return replayMsg
}

// /shell2telegram rm "/command" - delete command
func cmdShell2telegramRm(ctx Ctx) (replayMsg string) {
	commandName := ctx.messageArgs

	if commandName == "" {
		return "Please set command for delete: /shell2telegram rm </command>"
	}
	if _, ok := ctx.commands[commandName]; ok {
		delete(ctx.commands, commandName)
		replayMsg = "Deleted command: " + commandName
	} else {
		replayMsg = fmt.Sprintf("Command %s not found", commandName)
	}

	return replayMsg
}

// /shell2telegram version - get version
func cmdShell2telegramVersion(_ Ctx) (replayMsg string) {
	replayMsg = fmt.Sprintf("shell2telegram %s", version)
	return replayMsg
}

// /shell2telegram exit - terminate bot
func cmdShell2telegramExit(ctx Ctx) (replayMsg string) {
	if ctx.appConfig.addExit {
		replayMsg = "bye..."
		go func() {
			ctx.exitSignal <- struct{}{}
		}()
	}
	return replayMsg
}

// /shell2telegram broadcast_to_root - broadcast message to root users in private chat
func cmdShell2telegramBroadcastToRoot(ctx Ctx) (replayMsg string) {
	message := ctx.messageArgs

	if message == "" {
		replayMsg = "Please set message: /shell2telegram broadcast_to_root <message>"
	} else {
		ctx.users.BroadcastForRoots(ctx.messageSignal,
			fmt.Sprintf("Message from %s:\n%s", ctx.users.String(ctx.userID), message),
			ctx.userID, // don't send self
		)
		replayMsg = "Message sent"
	}

	return replayMsg
}

// /shell2telegram message_to_user user_id|username "message" - send message to user in private chat
func cmdShell2telegramMessageToUser(ctx Ctx) (replayMsg string) {
	userName, message := splitStringHalfBySpace(ctx.messageArgs)

	if userName == "" || message == "" {
		replayMsg = "Please set user_name and message: /shell2telegram message_to_user <user_id|username> <message>"
	} else {
		userID := ctx.users.FindByIDOrUserName(userName)

		if userID > 0 {
			ctx.users.SendMessageToPrivate(ctx.messageSignal, userID, message)
			replayMsg = "Message sent"
		} else {
			replayMsg = "User not found"
		}
	}

	return replayMsg
}

// cmdHistory - show command history for user
func cmdHistory(ctx Ctx) (replayMsg string) {
	user, ok := ctx.users.list[ctx.userID]
	if !ok {
		return "User not found"
	}

	subCmd := strings.TrimSpace(ctx.messageArgs)

	// Export history to file
	if subCmd == "export" {
		if len(user.History) == 0 {
			return "No command history to export"
		}

		var historyText string
		historyText = fmt.Sprintf("Command History for %s\n", ctx.users.String(ctx.userID))
		historyText += fmt.Sprintf("Total entries: %d\n\n", len(user.History))

		for i, entry := range user.History {
			status := "✓"
			if !entry.Success {
				status = "✗"
			}
			historyText += fmt.Sprintf("%d. [%s] %s %s %s\n   %s\n\n",
				i+1,
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				status,
				entry.Command,
				entry.Arguments,
				"---",
			)
		}

		// Save to temp file
		fileName := fmt.Sprintf("history_%d_%d.txt", ctx.userID, time.Now().Unix())
		filePath := TempFilesDir + string(os.PathSeparator) + fileName
		err := ioutil.WriteFile(filePath, []byte(historyText), 0644)
		if err != nil {
			return fmt.Sprintf("Error creating history file: %s", err)
		}

		// Return file path with FILE: prefix to trigger file sending
		return "FILE:" + filePath
	}

	// Show last 10 commands with inline buttons to repeat
	if len(user.History) == 0 {
		return "No command history yet"
	}

	replayMsg = "Recent command history:\n\n"

	// Get last 10 entries
	start := 0
	if len(user.History) > 10 {
		start = len(user.History) - 10
	}

	for i := start; i < len(user.History); i++ {
		entry := user.History[i]
		status := "✓"
		if !entry.Success {
			status = "✗"
		}
		replayMsg += fmt.Sprintf("%d. [%s] %s %s %s\n",
			i-start+1,
			entry.Timestamp.Format("15:04:05"),
			status,
			entry.Command,
			entry.Arguments,
		)
	}

	return replayMsg
}

// /shell2telegram logs - show recent command logs from database
func cmdShell2telegramLogs(ctx Ctx) (replayMsg string) {
	if ctx.database == nil {
		return "Database logging is not enabled. Use -enable-db-logging flag"
	}

	limit := 20
	if ctx.messageArgs != "" {
		fmt.Sscanf(ctx.messageArgs, "%d", &limit)
	}
	if limit > 100 {
		limit = 100
	}

	logs, err := ctx.database.GetRecentLogs(limit)
	if err != nil {
		return fmt.Sprintf("Error fetching logs: %v", err)
	}

	if len(logs) == 0 {
		return "No logs found"
	}

	replayMsg = fmt.Sprintf("Recent %d command logs:\n\n", len(logs))
	for _, log := range logs {
		status := "✓"
		if !log.Success {
			status = "✗"
		}
		replayMsg += fmt.Sprintf("[%s] %s %s\n@%s: %s %s\nTime: %dms\n\n",
			log.Timestamp.Format("2006-01-02 15:04"),
			status,
			log.Command,
			log.Username,
			log.Command,
			log.Arguments,
			log.ExecutionTime,
		)
	}

	return replayMsg
}

// /shell2telegram search_logs - search logs by keyword
func cmdShell2telegramSearchLogs(ctx Ctx) (replayMsg string) {
	if ctx.database == nil {
		return "Database logging is not enabled. Use -enable-db-logging flag"
	}

	keyword := ctx.messageArgs
	if keyword == "" {
		return "Usage: /shell2telegram search_logs <keyword>"
	}

	logs, err := ctx.database.SearchLogs(keyword, 20)
	if err != nil {
		return fmt.Sprintf("Error searching logs: %v", err)
	}

	if len(logs) == 0 {
		return fmt.Sprintf("No logs found matching '%s'", keyword)
	}

	replayMsg = fmt.Sprintf("Found %d logs matching '%s':\n\n", len(logs), keyword)
	for _, log := range logs {
		status := "✓"
		if !log.Success {
			status = "✗"
		}
		replayMsg += fmt.Sprintf("[%s] %s %s\n@%s: %s %s\n\n",
			log.Timestamp.Format("2006-01-02 15:04"),
			status,
			log.Command,
			log.Username,
			log.Command,
			log.Arguments,
		)
	}

	return replayMsg
}

// /shell2telegram db_stats - show database statistics
func cmdShell2telegramDBStats(ctx Ctx) (replayMsg string) {
	if ctx.database == nil {
		return "Database logging is not enabled. Use -enable-db-logging flag"
	}

	stats, err := ctx.database.GetStats()
	if err != nil {
		return fmt.Sprintf("Error fetching stats: %v", err)
	}

	replayMsg = "Database Statistics:\n\n"
	replayMsg += fmt.Sprintf("Total commands: %d\n", stats["total_commands"])
	replayMsg += fmt.Sprintf("Successful: %d\n", stats["successful_commands"])
	replayMsg += fmt.Sprintf("Success rate: %.2f%%\n", stats["success_rate"])
	replayMsg += fmt.Sprintf("Avg execution time: %.2f ms\n", stats["avg_execution_time_ms"])

	if mostUsed, ok := stats["most_used_command"]; ok {
		replayMsg += fmt.Sprintf("\nMost used command: %s (%d times)\n", mostUsed, stats["most_used_count"])
	}

	return replayMsg
}

// Note: Inline keyboard feature (generateInlineKeyboard) requires telegram-bot-api v5+
// This is left as a future enhancement when the library is upgraded
