package main

import (
    "encoding/json"
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

type AddGroupRequest struct {
    Link string `json:"link"`
}

func main() {
    // Переменные для бд
    dbHost := getEnv("DB_HOST", "localhost")
    dbUser := getEnv("DB_USER", "postgres")
    dbPassword := getEnv("DB_PASSWORD", "root")
    dbName := getEnv("DB_NAME", "telescrap")
    dbPort := getEnv("DB_PORT", "5430")
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

    // Переменные для тг
    appID := os.Getenv("APP_ID")
    appHash := os.Getenv("APP_HASH")
    phone := os.Getenv("USER_PHONE")

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

    // Загружаем html шаблоны
    tmplPath := filepath.Join(".", "telescrap", "templates", "pages")
    if err := templates.LoadTemplates(tmplPath); err != nil {
        log.Fatalf("Ошибка загрузки шаблонов: %v", err)
    }

    go startWebServer()

    // Хендлер сообщений для телеги
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
        messageData.TGID = strconv.FormatInt(int64(message.ID), 10)
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

        if !IsHere(messageData.SenderID, db) {
            senderIDInt, _ := strconv.ParseInt(messageData.SenderID, 10, 64)
            if err := AddUser(client, db, senderIDInt); err != nil {
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
// тут типаAPI
func startWebServer() {
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/api/users", apiUsersHandler)
    http.HandleFunc("/api/groups", apiGroupsHandler)
    http.HandleFunc("/api/messages", apiMessagesHandler)
    http.HandleFunc("/api/join", apiJoinGroupHandler)

    log.Println("Веб-интерфейс запущен на :6767")
    if err := http.ListenAndServe(":6767", nil); err != nil {
        log.Fatal("Ошибка веб-сервера: ", err)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    if err := templates.Render(w, "layout.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func apiUsersHandler(w http.ResponseWriter, r *http.Request) {
    users, err := database.Get[structs.User](time.Time{}, time.Time{}, db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func apiGroupsHandler(w http.ResponseWriter, r *http.Request) {
    groups, err := database.Get[structs.Group](time.Time{}, time.Time{}, db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(groups)
}

func apiMessagesHandler(w http.ResponseWriter, r *http.Request) {
    chatID := r.URL.Query().Get("chat_id")
    var messages []structs.Message
    
    query := db
    if chatID != "" {
        query = query.Where("group_id = ?", chatID)
    }

    if err := query.Find(&messages).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}

func apiJoinGroupHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req AddGroupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    err := AddByLink(req.Link)
    if err != nil {
        http.Error(w, fmt.Sprintf("Ошибка добавления: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func AddByLink(link string) error {
    log.Printf("Попытка добавить бота в группу: %s", link)
    return fmt.Errorf("функция AddByLink не реализована")
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

func IsHere(id string, db *gorm.DB) bool {
    var count int64
    db.Model(&structs.User{}).Where("ID = ?", id).Count(&count)
    return count > 0
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
    userStructed.ID = strconv.FormatInt(user.ID, 10)
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