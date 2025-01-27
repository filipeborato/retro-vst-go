package handlers
import (
    "context"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"

    //"retro-vst-go/domain"
    "gorm.io/gorm"
)

var googleOAuthConfig = &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  os.Getenv("URL_CALLBACK"),
    Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
    Endpoint:     google.Endpoint,
}

func GoogleCallbackHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        code := c.Query("code")
        if code == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
            return
        }

        // Troca o code por token
        token, err := googleOAuthConfig.Exchange(context.Background(), code)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token"})
            return
        }

        // Usa esse token para buscar info do usuário no endpoint do Google
        client := googleOAuthConfig.Client(context.Background(), token)
        resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
            return
        }
        defer resp.Body.Close()

        // Parse do JSON (name, email, id, etc.)
        // ...
        // googleUser := struct {
        //    Id    string `json:"id"`
        //    Email string `json:"email"`
        //    Name  string `json:"name"`
        // }{}
        // json.NewDecoder(resp.Body).Decode(&googleUser)

        // Verifica se já existe User com google_id = googleUser.Id
        //var user domain.User
        // if err := db.Where("google_id = ?", googleUser.Id).First(&user).Error; err != nil {
        //    // se não existir, criar
        //    user = domain.User{
        //       Name: googleUser.Name,
        //       Email: strings.ToLower(googleUser.Email),
        //       GoogleID: &googleUser.Id,
        //    }
        //    db.Create(&user)
        // }

        // Gera JWT
        // tokenString, err := CreateJWT(user)
        // if err != nil {
        //    ...
        // }

        // Cria session com AuthMethod = "google"
        // session := domain.Session{
        //   UserID: user.ID,
        //   AuthMethod: "google",
        // }
        // db.Create(&session)

        // c.JSON(http.StatusOK, gin.H{
        //    "token": tokenString,
        //    "message": "Login via Google successful",
        // })
        c.JSON(http.StatusOK, gin.H{"message": "Exemplo de callback - implementar parse user info"})
    }
}
