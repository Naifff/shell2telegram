package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/msoap/raphanus"
	tgbotapi "gopkg.in/telegram-bot-api.v2"
)

const (
	// version - current version
	version = "1.10"

	// DefaultBotTimeout - bot default timeout
	DefaultBotTimeout = 60

	// MessagesQueueSize - size of channel for bot messages
	MessagesQueueSize = 10

	// MaxMessageLength - max length of one bot message
	MaxMessageLength = 4096

	// SecondsForAutoSaveUsersToDB - save users to file every 1 min (if need)
	SecondsForAutoSaveUsersToDB = 60

	// DBFileName - DB json name
	DBFileName = "shell2telegram.json"

	// TempFilesDir - temporary directory for uploaded files
	TempFilesDir = "/tmp/shell2telegram_files"

	// MaxFileAge - max age for temp files in seconds (1 hour)
	MaxFileAge = 3600

	// shell2telegram command name for get plain text without /command
	cmdPlainText = "/:plain_text"
)

// Command - one user command
type Command struct {
	shellCmd    string   // shell command
	description string   // command description for list in /help (/cmd:desc="Command name")
	vars        []string // environment vars for user text, split by `/s+` to vars (/cmd:vars=SUBCOMMAND,ARGS)
	isMarkdown  bool     // send message in markdown format
	isFile      bool     // command output is a file path - send file to user
	// buttons - inline buttons feature requires telegram-bot-api v5+ (future enhancement)
}

// Commands - list of all commands
type Commands map[string]Command

// Config - config struct
type Config struct {
	token                  string   // bot token
	botTimeout             int      // bot timeout
	predefinedAllowedUsers []string // telegram users who are allowed to chat with the bot
	predefinedRootUsers    []string // telegram users, who confirms new users in their private chat
	description            string   // description of bot
	bindAddr               string   // bind address to listen webhook requests
	webhookURL             url.URL  // url for the webhook
	usersDB                string   // file for store users
	shell                  string   // custom shell
	cache                  int      // caching command out (in seconds)
	shTimeout              int      // timeout for execute shell command (in seconds)
	proxyServer            string   // proxy server address (host:port or http://host:port)
	proxyUser              string   // proxy username
	proxyPassword          string   // proxy password
	addExit                bool     // adding /shell2telegram exit command
	allowAll               bool     // allow all user (DANGEROUS!)
	logCommands            bool     // logging all commands
	persistentUsers        bool     // load/save users from file
	isPublicBot            bool     // bot is public (don't add /auth* commands)
	oneThread              bool     // run each shell commands in one thread
	enableDBLogging        bool     // enable database logging
	dbPath                 string   // path to SQLite database
}

// message types
const (
	msgIsText int8 = iota
	msgIsPhoto
	msgIsDocument
)

// BotMessage - record for send via channel for send message to telegram chat
type BotMessage struct {
	message     string
	fileName    string
	filePath    string // path to file for sending as document
	photo       []byte
	chatID      int
	messageType int8
	isMarkdown  bool
	// keyboard - inline keyboard requires telegram-bot-api v5+ (future enhancement)
}

// ----------------------------------------------------------------------------
// get config
func getConfig() (commands Commands, appConfig Config, err error) {
	flag.StringVar(&appConfig.token, "tb-token", "", "setting bot `token` (or set TB_TOKEN variable)")
	flag.BoolVar(&appConfig.addExit, "add-exit", false, "adding \"/shell2telegram exit\" command for terminate bot (for roots only)")
	flag.IntVar(&appConfig.botTimeout, "timeout", DefaultBotTimeout, "setting timeout for bot (in `seconds`)")
	flag.StringVar(&appConfig.bindAddr, "bind-addr", "", "bind address to listen webhook requests, like: `0.0.0.0:8080`")
	flag.Var(&urlValue{&appConfig.webhookURL}, "webhook", "`url` of bot's webhook")
	flag.BoolVar(&appConfig.allowAll, "allow-all", false, "allow all users (DANGEROUS!)")
	flag.BoolVar(&appConfig.logCommands, "log-commands", false, "logging all commands")
	flag.StringVar(&appConfig.description, "description", "", "setting description of bot")
	flag.BoolVar(&appConfig.persistentUsers, "persistent-users", false, "load/save users from file (default ~/.config/shell2telegram.json)")
	flag.StringVar(&appConfig.usersDB, "users-db", "", "`file` for store users")
	flag.IntVar(&appConfig.cache, "cache", 0, "caching command out (in `seconds`)")
	flag.BoolVar(&appConfig.isPublicBot, "public", false, "bot is public (don't add /auth* commands)")
	flag.IntVar(&appConfig.shTimeout, "sh-timeout", 0, "set timeout for execute shell command (in `seconds`)")
	flag.StringVar(&appConfig.shell, "shell", "sh", "custom shell or \"\" for execute without shell")
	flag.BoolVar(&appConfig.oneThread, "one-thread", false, "run each shell command in one thread")
	flag.StringVar(&appConfig.proxyServer, "proxy-server", "", "proxy server `address` (host:port or http://host:port)")
	flag.StringVar(&appConfig.proxyUser, "proxy-user", "", "proxy `username`")
	flag.StringVar(&appConfig.proxyPassword, "proxy-password", "", "proxy `password`")
	flag.BoolVar(&appConfig.enableDBLogging, "enable-db-logging", false, "enable SQLite database logging of all commands")
	flag.StringVar(&appConfig.dbPath, "db-path", "", "`path` to SQLite database file (default ~/.config/shell2telegram.db)")
	logFilename := flag.String("log", "", "log `filename`, default - STDOUT")
	predefinedAllowedUsers := flag.String("allow-users", "", "telegram users who are allowed to chat with the bot (\"user1,user2\")")
	predefinedRootUsers := flag.String("root-users", "", "telegram users, who confirms new users in their private chat (\"user1,user2\")")
	showVersion := flag.Bool("version", false, "get version")

	flag.Usage = func() {
		fmt.Printf("usage: %s [options] %s\n%s\n%s\n\noptions:\n",
			os.Args[0],
			`/chat_command "shell command" /chat_command2 "shell command2"`,
			"All text after /chat_command will be sent to STDIN of shell command.",
			"If chat command is /:plain_text - get user message without any /command (for private chats only)",
		)
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// setup log file
	if len(*logFilename) > 0 {
		fhLog, err := os.OpenFile(*logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
		log.SetOutput(fhLog)
	}

	// setup users and roots
	if *predefinedAllowedUsers != "" {
		appConfig.predefinedAllowedUsers = strings.Split(*predefinedAllowedUsers, ",")
	}
	if *predefinedRootUsers != "" {
		appConfig.predefinedRootUsers = strings.Split(*predefinedRootUsers, ",")
	}

	commands = Commands{}
	// need >= 2 arguments and count of it must be even
	args := flag.Args()
	if len(args) < 2 || len(args)%2 == 1 {
		return commands, appConfig, fmt.Errorf("error: need pairs of /chat-command and shell-command")
	}

	for i := 0; i < len(args); i += 2 {
		path, command, err := parseBotCommand(args[i], args[i+1]) // (/path, shell_command)
		if err != nil {
			return commands, appConfig, err
		}
		commands[path] = command
	}

	// setup proxy from environment variables if not set via flags
	if appConfig.proxyServer == "" {
		// Try custom environment variables first
		if proxyEnv := os.Getenv("PROXY_SERVER"); proxyEnv != "" {
			appConfig.proxyServer = proxyEnv
		} else if proxyEnv := os.Getenv("HTTP_PROXY"); proxyEnv != "" {
			// Try standard HTTP_PROXY variable (uppercase)
			appConfig.proxyServer = proxyEnv
		} else if proxyEnv := os.Getenv("http_proxy"); proxyEnv != "" {
			// Try standard http_proxy variable (lowercase)
			appConfig.proxyServer = proxyEnv
		} else if proxyEnv := os.Getenv("HTTPS_PROXY"); proxyEnv != "" {
			// Try standard HTTPS_PROXY variable
			appConfig.proxyServer = proxyEnv
		} else if proxyEnv := os.Getenv("https_proxy"); proxyEnv != "" {
			// Try standard https_proxy variable (lowercase)
			appConfig.proxyServer = proxyEnv
		}
	}

	// setup proxy credentials from environment variables if not set via flags
	if appConfig.proxyUser == "" {
		if proxyUser := os.Getenv("PROXY_USER"); proxyUser != "" {
			appConfig.proxyUser = proxyUser
		}
	}
	if appConfig.proxyPassword == "" {
		if proxyPass := os.Getenv("PROXY_PASSWORD"); proxyPass != "" {
			appConfig.proxyPassword = proxyPass
		}
	}

	if appConfig.token == "" {
		if appConfig.token = os.Getenv("TB_TOKEN"); appConfig.token == "" {
			return commands, appConfig, fmt.Errorf("TB_TOKEN environment var not found. See https://core.telegram.org/bots#botfather for more information")
		}
	}

	return commands, appConfig, nil
}

// ----------------------------------------------------------------------------
func sendMessage(messageSignal chan<- BotMessage, chatID int, message []byte, isMarkdown bool, isFileOutput bool) {
	go func() {
		messageString := string(message)

		// Check if output is a file path (for :file modifier or FILE: prefix)
		if isFileOutput || strings.HasPrefix(messageString, "FILE:") {
			filePath := messageString
			if strings.HasPrefix(messageString, "FILE:") {
				filePath = strings.TrimSpace(strings.TrimPrefix(messageString, "FILE:"))
			}

			// Send as document
			messageSignal <- BotMessage{
				chatID:      chatID,
				messageType: msgIsDocument,
				filePath:    filePath,
			}
			return
		}

		var fileName string
		fileType := http.DetectContentType(message)
		switch fileType {
		case "image/png":
			fileName = "file.png"
		case "image/jpeg":
			fileName = "file.jpeg"
		case "image/gif":
			fileName = "file.gif"
		case "image/bmp":
			fileName = "file.bmp"
		case "video/mp4":
			// TODO: nedded migrate to new telegram-bot-api library
			log.Printf("not supported")
			return
		default:
			fileName = "message"
		}

		if fileName == "message" {
			// is text message
			var messagesList []string

			if len(messageString) <= MaxMessageLength {
				messagesList = []string{messageString}
			} else {
				messagesList = splitStringLinesBySize(messageString, MaxMessageLength)
			}

			for _, messageChunk := range messagesList {
				messageSignal <- BotMessage{
					chatID:      chatID,
					messageType: msgIsText,
					message:     messageChunk,
					isMarkdown:  isMarkdown,
				}
			}

		} else {
			// is image
			messageSignal <- BotMessage{
				chatID:      chatID,
				messageType: msgIsPhoto,
				fileName:    fileName,
				photo:       message,
			}
		}
	}()
}

// ----------------------------------------------------------------------------
// createBotWithProxy - create bot with optional proxy support
func createBotWithProxy(token string, config *Config) (*tgbotapi.BotAPI, error) {
	if config.proxyServer == "" {
		// No proxy configured, use default client
		return tgbotapi.NewBotAPI(token)
	}

	// Configure proxy
	proxyURLStr := config.proxyServer
	if !strings.HasPrefix(proxyURLStr, "http://") && !strings.HasPrefix(proxyURLStr, "https://") {
		proxyURLStr = "http://" + proxyURLStr
	}

	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %v", err)
	}

	// Add authentication if provided via separate parameters
	// or if not present in URL already
	if proxyURL.User == nil && config.proxyUser != "" {
		// Use credentials from flags/env variables
		proxyURL.User = url.UserPassword(config.proxyUser, config.proxyPassword)
	}
	// If credentials are in URL (http://user:pass@proxy:8080), they are already parsed

	// Create HTTP client with proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	httpClient := &http.Client{
		Transport: transport,
	}

	// Create bot with custom HTTP client
	bot, err := tgbotapi.NewBotAPIWithClient(token, httpClient)
	if err != nil {
		return nil, err
	}

	log.Printf("Using proxy server: %s", config.proxyServer)
	return bot, nil
}

// ----------------------------------------------------------------------------
// Note: handleCallbackQuery for inline buttons requires telegram-bot-api v5+
// This feature is left as a future enhancement when the library is upgraded

// ----------------------------------------------------------------------------
func main() {
	commands, appConfig, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := createBotWithProxy(appConfig.token, &appConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on bot account: @%s", bot.Self.UserName)

	tgbotConfig := tgbotapi.NewUpdate(0)
	tgbotConfig.Timeout = appConfig.botTimeout
	var botUpdatesChan <-chan tgbotapi.Update
	var server *http.Server

	if appConfig.bindAddr != "" {
		_, err = bot.SetWebhook(tgbotapi.WebhookConfig{URL: &appConfig.webhookURL})
		if err != nil {
			log.Fatal(err)
		}

		botUpdatesChan = bot.ListenForWebhook(appConfig.webhookURL.Path)
		server = &http.Server{Addr: appConfig.bindAddr}
		go func() {
			log.Println("Listening incoming requests at ", appConfig.bindAddr)
			log.Fatal(server.ListenAndServe())
		}()
	} else {
		botUpdatesChan, err = bot.GetUpdatesChan(tgbotConfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	users := NewUsers(appConfig)
	messageSignal := make(chan BotMessage, MessagesQueueSize)
	vacuumTicker := time.Tick(SecondsForOldUsersBeforeVacuum * time.Second)
	saveToBDTicker := make(<-chan time.Time)
	oneThreadMutex := sync.Mutex{}
	exitSignal := make(chan struct{})
	systemExitSignal := make(chan os.Signal, 1)
	signal.Notify(systemExitSignal, os.Interrupt)

	if appConfig.persistentUsers {
		saveToBDTicker = time.Tick(SecondsForAutoSaveUsersToDB * time.Second)
	}

	var cache raphanus.DB
	if appConfig.cache > 0 {
		cache = raphanus.New()
	}

	// Initialize database if logging is enabled
	var database *Database
	if appConfig.enableDBLogging {
		database, err = InitDatabase(appConfig.dbPath)
		if err != nil {
			log.Printf("Failed to initialize database: %v. Continuing without DB logging.", err)
		} else {
			defer database.Close()
		}
	}

	// all /shell2telegram sub-commands handlers
	internalCommands := map[string]func(Ctx) string{
		"stat":              cmdShell2telegramStat,
		"ban":               cmdShell2telegramBan,
		"search":            cmdShell2telegramSearch,
		"desc":              cmdShell2telegramDesc,
		"rm":                cmdShell2telegramRm,
		"exit":              cmdShell2telegramExit,
		"version":           cmdShell2telegramVersion,
		"broadcast_to_root": cmdShell2telegramBroadcastToRoot,
		"message_to_user":   cmdShell2telegramMessageToUser,
		"logs":              cmdShell2telegramLogs,
		"search_logs":       cmdShell2telegramSearchLogs,
		"db_stats":          cmdShell2telegramDBStats,
	}

	// Create temp directory for uploaded files
	createDirIfNeed(TempFilesDir)

	// Cleanup old files ticker
	fileCleanupTicker := time.Tick(MaxFileAge * time.Second)

	doExit := false
	for !doExit {
		select {
		case telegramUpdate := <-botUpdatesChan:
			// Note: Callback query handling for inline buttons requires telegram-bot-api v5+ (future)

			// Skip if no message
			if telegramUpdate.Message.MessageID == 0 {
				continue
			}

			var messageCmd, messageArgs string
			allUserMessage := telegramUpdate.Message.Text

			// Handle document upload
			if telegramUpdate.Message.Document.FileID != "" {
				messageCmd = cmdPlainText // Will be processed as plain text if /:plain_text command exists
				allUserMessage = ""        // Document will be handled separately
			} else if len(allUserMessage) > 0 && allUserMessage[0] == '/' {
				messageCmd, messageArgs = splitStringHalfBySpace(allUserMessage)
			} else {
				messageCmd, messageArgs = cmdPlainText, allUserMessage
			}

			allowPlainText := false
			if _, ok := commands[cmdPlainText]; ok {
				allowPlainText = true
			}

			replayMsg := ""

			if len(messageCmd) > 0 && (messageCmd != cmdPlainText || allowPlainText) {

				users.AddNew(telegramUpdate.Message)
				userID := telegramUpdate.Message.From.ID
				allowExec := appConfig.allowAll || users.IsAuthorized(userID)

				// Download file if document was uploaded
				var uploadedFilePath string
				if telegramUpdate.Message.Document.FileID != "" && allowExec {
					filePath, err := downloadFile(bot, telegramUpdate.Message.Document.FileID, TempFilesDir)
					if err != nil {
						log.Printf("Failed to download file: %s", err)
						sendMessage(messageSignal, telegramUpdate.Message.Chat.ID, []byte(fmt.Sprintf("Error downloading file: %s", err)), false, false)
						continue
					}
					uploadedFilePath = filePath
					log.Printf("File uploaded by %s: %s", users.String(userID), filePath)
				}

				ctx := Ctx{
					appConfig:      &appConfig,
					users:          &users,
					commands:       commands,
					userID:         userID,
					allowExec:      allowExec,
					messageCmd:     messageCmd,
					messageArgs:    messageArgs,
					messageSignal:  messageSignal,
					chatID:         telegramUpdate.Message.Chat.ID,
					exitSignal:     exitSignal,
					cache:          &cache,
					oneThreadMutex: &oneThreadMutex,
					uploadedFile:   uploadedFilePath,
					database:       database,
				}

				switch {
				// commands .................................
				case !appConfig.isPublicBot && (messageCmd == "/auth" || messageCmd == "/authroot"):
					replayMsg = cmdAuth(ctx)

				case messageCmd == "/help":
					replayMsg = cmdHelp(ctx)

				case allowExec && messageCmd == "/history":
					replayMsg = cmdHistory(ctx)

				case messageCmd == "/shell2telegram" && users.IsRoot(userID):
					var messageSubCmd string
					messageSubCmd, messageArgs = splitStringHalfBySpace(messageArgs)
					ctx.messageArgs = messageArgs
					if cmdHandler, ok := internalCommands[messageSubCmd]; ok {
						replayMsg = cmdHandler(ctx)
					} else {
						replayMsg = "Sub-command not found"
					}

				case allowExec && (allowPlainText && messageCmd == cmdPlainText || messageCmd[0] == '/'):
					// Record command execution
					users.AddToHistory(userID, messageCmd, messageArgs, true)
					cmdUser(ctx)

				} // switch for commands

				if appConfig.logCommands {
					log.Printf("%s: %s", users.String(userID), allUserMessage)
				}

				sendMessage(messageSignal, telegramUpdate.Message.Chat.ID, []byte(replayMsg), false, false)
			}

		case botMessage := <-messageSignal:
			switch {
			case botMessage.messageType == msgIsText && !stringIsEmpty(botMessage.message):
				messageConfig := tgbotapi.NewMessage(botMessage.chatID, botMessage.message)
				if botMessage.isMarkdown {
					messageConfig.ParseMode = tgbotapi.ModeMarkdown
				}
				// Note: Inline keyboard support requires telegram-bot-api v5+
				_, err = bot.Send(messageConfig)
			case botMessage.messageType == msgIsPhoto && len(botMessage.photo) > 0:
				bytesPhoto := tgbotapi.FileBytes{Name: botMessage.fileName, Bytes: botMessage.photo}
				_, err = bot.Send(tgbotapi.NewPhotoUpload(botMessage.chatID, bytesPhoto))
			case botMessage.messageType == msgIsDocument && botMessage.filePath != "":
				// Send file as document
				_, err = bot.Send(tgbotapi.NewDocumentUpload(botMessage.chatID, botMessage.filePath))
				if err != nil {
					// If file sending failed, try to send error message
					log.Printf("failed to send document %s: %s", botMessage.filePath, err)
					errorMsg := fmt.Sprintf("Error: cannot send file '%s': %s", botMessage.filePath, err)
					bot.Send(tgbotapi.NewMessage(botMessage.chatID, errorMsg))
				}
			}

			if err != nil {
				log.Printf("failed to send message: %s", err)
			}

		case <-fileCleanupTicker:
			cleanupOldFiles(TempFilesDir, MaxFileAge)

		case <-saveToBDTicker:
			users.SaveToDB(appConfig.usersDB)

		case <-vacuumTicker:
			users.ClearOldUsers()

		case <-systemExitSignal:
			go func() {
				exitSignal <- struct{}{}
			}()

		case <-exitSignal:
			if appConfig.persistentUsers {
				users.needSaveDB = true
				users.SaveToDB(appConfig.usersDB)
			}
			if server != nil {
				log.Println(server.Close())
			}
			doExit = true
		}
	}
}
