package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "path/filepath"
    "strconv"
    "syscall"
    "telescrap/storage/database"
    "telescrap/structs"
    "telescrap/templates"
    "time"

    "github.com/amarnathcjd/gogram/telegram"
    "github.com/google/uuid"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var db *gorm.DB
var tgClient *telegram.Client

// PageData для передачи данных в шаблон
type PageData struct {
    Users        []structs.User
    Groups       []structs.Group
    Messages     []structs.Message
    Links        []structs.Link
    MediaFiles   []structs.Message
    ChatList     []templates.ChatItem // Новый тип для списка чатов
    SelectedChat string
    CurrentPanel string
}

func main() {
    // Переменные для бд
    dbHost := getEnv("DB_HOST", "localhost")
    dbUser := getEnv("DB_USER", "postgres")
    dbPassword := getEnv("DB_PASSWORD", "postgres")
    dbName := getEnv("DB_NAME", "telescrap")
    dbPort := getEnv("DB_PORT", "5432")
    dbSSLMode := getEnv("DB_SSLMODE", "disable")

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
        dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        log.Fatalf("failed to connect database: %v", err)
    }

    if err := database.Init(db); err != nil {
        log.Fatalf("failed to init database: %v", err)
    }

    // Переменные для тг
    appID := os.Getenv("TG_APP_ID")
    appHash := ("TG_APP_HASH")
    phone := os.Getenv("TG_APP_PHONE")

    appIDInt, err := strconv.Atoi(appID)
    if err != nil {
        log.Fatal("Неверный APP_ID:", err)
    }

    client, err := telegram.NewClient(telegram.ClientConfig{
        AppID:   int32(appIDInt),
        AppHash: appHash,
    })
    if err != nil {
        log.Fatal("Ошибка создания клиента:", err)
    }
    tgClient = client

    client.Conn()

    _, err = client.Login(phone)
    if err != nil {
        log.Fatal("Ошибка авторизации:", err)
    }
    log.Println("Бот успешно авторизован!")

    tmplPath := filepath.Join(".", "templates", "pages")
    if err := templates.LoadTemplates(tmplPath); err != nil {
        log.Fatalf("Ошибка загрузки шаблонов: %v", err)
    }

    go startWebServer()

    // Хендлер сообщений
    client.On(telegram.OnMessage, func(message *telegram.NewMessage) error {
        me, _ := client.GetMe()
        if me != nil && message.SenderID() == me.ID {
            return nil
        }

        var mediaPaths []string
        var savedFilePath string

        if message.Photo() != nil {
            photo := message.Photo()
            dir := filepath.Join("files", fmt.Sprint(message.ID))
            if err := os.MkdirAll(dir, 0755); err != nil {
                log.Printf("Ошибка создания папки: %v", err)
            } else {
                fileName := fmt.Sprintf("photo_%d.jpg", photo.ID)
                filePath := filepath.Join(dir, fileName)
                if _, err := client.DownloadMedia(photo, &telegram.DownloadOptions{FileName: filePath}); err != nil {
                    log.Printf("Ошибка скачивания фото: %v", err)
                } else {
                    savedFilePath = filePath
                }
            }
        }

        if message.Document() != nil {
            doc := message.Document()
            dir := filepath.Join("files", fmt.Sprint(message.ID))
            if err := os.MkdirAll(dir, 0755); err != nil {
                log.Printf("Ошибка создания папки: %v", err)
            } else {
                fileName := fmt.Sprintf("file_%d.bin", doc.ID)
                for _, attr := range doc.Attributes {
                    if filenameAttr, ok := attr.(*telegram.DocumentAttributeFilename); ok {
                        fileName = filenameAttr.FileName
                        break
                    }
                }
                if filepath.Ext(fileName) == "" {
                    fileName += getExtensionFromMime(doc.MimeType)
                }
                filePath := filepath.Join(dir, fileName)
                if _, err := client.DownloadMedia(doc, &telegram.DownloadOptions{FileName: filePath}); err != nil {
                    log.Printf("Ошибка скачивания документа: %v", err)
                } else {
                    savedFilePath = filePath
                }
            }
        }

        if savedFilePath != "" {
            mediaPaths = append(mediaPaths, savedFilePath)
        }

        var messageData structs.Message
        messageData.ID = uuid.New().String()
        messageData.TGID = int(int64(message.ID))
        messageData.SenderID = strconv.FormatInt(message.SenderID(), 10)
        messageData.GroupID = strconv.FormatInt(message.ChatID(), 10)
        messageData.Message = message.Text()

        if message.IsReply() {
            if replyID := message.ReplyToMsgID(); replyID != 0 {
                messageData.ReplyTo = strconv.FormatInt(int64(replyID), 10)
            }
        }

        messageData.IsPinned = message.Message.Pinned
        messageData.Media = mediaPaths

        senderIDInt, _ := strconv.Atoi(messageData.SenderID)
        if !database.IsHere(senderIDInt, db) {
            senderIDInt64, _ := strconv.ParseInt(messageData.SenderID, 10, 64)
            if err := AddUser(client, db, senderIDInt64); err != nil {
                log.Printf("Ошибка добавления юзера в БД: %v", err)
            }
        }

        if err := database.Add(db, &messageData); err != nil {
            log.Printf("Ошибка сохранения в БД: %v", err)
            return err
        }

        log.Printf("Сообщение %d сохранено в БД", message.ID)
        return nil
    })

    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

    log.Println("Бот запущен и ожидает сообщений...")
    <-sc

    log.Println("Получен сигнал остановки, завершаем работу...")
    client.Disconnect()
}

func startWebServer() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
    http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./files"))))

    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/join", joinGroupHandler)

    log.Println("Веб-интерфейс запущен на :6767")
    if err := http.ListenAndServe(":6767", nil); err != nil {
        log.Fatal("Ошибка веб-сервера: ", err)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    selectedChat := r.URL.Query().Get("chat_id")
    currentPanel := r.URL.Query().Get("panel")
    if currentPanel == "" {
        currentPanel = "chats"
    }

    users, _ := database.Get[structs.User](db)
    groups, _ := database.Get[structs.Group](db)
    links, _ := database.Get[structs.Link](db)

    var messages []structs.Message
    if selectedChat != "" {
        messages, _ = database.Get[structs.Message](db, database.WithGroupID(selectedChat))
    }

    var mediaFiles []structs.Message
    db.Where("media IS NOT NULL").Find(&mediaFiles)

    chatList := make([]templates.ChatItem, 0)
    
    for _, g := range groups {
        chatList = append(chatList, templates.ChatItem{
            ID:          g.TGID,
            Name:        g.Name,
            Avatar:      string([]rune(g.Name)[0]), 
            IsGroup:     true,
            Description: g.Description,
            Subscribers: fmt.Sprintf("%d", g.Subscribers),
        })
    }

    // Создаем мапу групп для быстрого поиска
    groupMap := make(map[string]bool)
    for _, g := range groups {
        groupMap[g.TGID] = true
    }
    
    // Создаем мапу юзеров для поиска по ID
    userMap := make(map[string]structs.User)
    for _, u := range users {
        userMap[strconv.Itoa(u.ID)] = u
    }

    var allMessages []structs.Message
    db.Select("group_id").Find(&allMessages)
    
    seenChats := make(map[string]bool) 
    
    for _, msg := range allMessages {
        if msg.GroupID == "" || seenChats[msg.GroupID] {
            continue
        }
        seenChats[msg.GroupID] = true
        
        // Если это не группа, пробуем найти юзера
        if !groupMap[msg.GroupID] {
            if user, ok := userMap[msg.GroupID]; ok {
                name := user.FirstName
                if name == "" {
                    name = user.Username
                }
                avatar := "?"
                if len(name) > 0 {
                    avatar = string([]rune(name)[0])
                }
                chatList = append(chatList, templates.ChatItem{
                    ID:          msg.GroupID,
                    Name:        name,
                    Avatar:      avatar,
                    IsGroup:     false,
                    Description: user.Username,
                    Subscribers: "Личный чат",
                })
            }
        }
    }

    data := PageData{
        Users:        users,
        Groups:       groups,
        Messages:     messages,
        Links:        links,
        MediaFiles:   mediaFiles,
        ChatList:     chatList, 
        SelectedChat: selectedChat,
        CurrentPanel: currentPanel,
    }

    if err := templates.Render(w, "layout.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func joinGroupHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    link := r.FormValue("link")
    if link != "" {
        err := AddByLink(link)
        if err != nil {
            log.Printf("Ошибка добавления группы: %v", err)
        }
    }
    http.Redirect(w, r, "/?panel=groups", http.StatusSeeOther)
}

func AddByLink(link string) error {
    log.Printf("Попытка добавить бота в группу: %s", link)
    return nil
}

func getEnv(env string, def string) string {
    val := os.Getenv(env)
    if val == "" {
        return def
    }
    return val
}

func getExtensionFromMime(mime string) string {
    switch mime {
    case "image/jpeg", "image/jpg":
        return ".jpg"
    case "image/png":
        return ".png"
    case "image/gif":
        return ".gif"
    case "image/webp":
        return ".webp"
    case "video/mp4":
        return ".mp4"
    case "audio/mpeg":
        return ".mp3"
    case "application/pdf":
        return ".pdf"
    default:
        return ".bin"
    }
}

func AddUser(client *telegram.Client, db *gorm.DB, id int64) error {
    resolvedPeer, err := client.ResolvePeer(id)
    if err != nil {
        return fmt.Errorf("ошибка ResolvePeer: %w", err)
    }

    inputPeerUser, ok := resolvedPeer.(*telegram.InputPeerUser)
    if !ok {
        return fmt.Errorf("объект не является пользователем")
    }

    inputUser := &telegram.InputUserObj{
        UserID:     inputPeerUser.UserID,
        AccessHash: inputPeerUser.AccessHash,
    }

    fullUser, err := client.UsersGetFullUser(inputUser)
    if err != nil {
        return fmt.Errorf("ошибка UsersGetFullUser: %w", err)
    }

    if len(fullUser.Users) == 0 {
        return fmt.Errorf("список пользователей пуст")
    }

    user, ok := fullUser.Users[0].(*telegram.UserObj)
    if !ok {
        return fmt.Errorf("не удалось преобразовать тип пользователя")
    }

    var userStructed structs.User
    userStructed.ID = int(user.ID)
    userStructed.Username = user.Username
    userStructed.PhoneNumber = user.Phone
    userStructed.FirstName = user.FirstName
    userStructed.LastName = user.LastName
    userStructed.CreatedAt = time.Now()

    if fullUser.FullUser != nil {
        userStructed.Description = fullUser.FullUser.About
    }

    if err := database.Add(db, &userStructed); err != nil {
        return fmt.Errorf("ошибка сохранения юзера в БД: %w", err)
    }

    log.Printf("Новый пользователь %d сохранен", id)
    return nil
}