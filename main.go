package main

import (
    "log"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/appleboy/gin-jwt/v2"
    "github.com/gin-contrib/cors"
)

func main() {
    jwtMiddleware, err := newJwtMiddleware()
    if err != nil {
        log.Fatalf("JWT middleware initialization error: %v", err)
    }
    r := gin.Default()

    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello from Go backend!"})
    })

    r.GET("/health_check", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.NoRoute(func(c *gin.Context) {
        c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
    })

    api := r.Group("/api")
    {
        api.POST("/login", jwtMiddleware.LoginHandler)
        api.GET("/refresh_token", jwtMiddleware.RefreshHandler)

        me := api.Group("/users/me").Use(jwtMiddleware.MiddlewareFunc())
        {
            me.GET("", func(c *gin.Context) {
                userID := userIdInJwt(c)
                // TODO : 一般的にはデータベースやストレージ、SaaSからuserIDを元にユーザー情報を取得する
                c.JSON(http.StatusOK, gin.H{
                    "userID": userID,
                })
            })
        }
    }

    port := "8080"

    if err := http.ListenAndServe(":"+port, r); err != nil {
        log.Fatal(err)
    }
}

func userIdInJwt(c *gin.Context) string {
    claims := jwt.ExtractClaims(c)
    userID := claims[jwt.IdentityKey]
    return userID.(string)
}

func newJwtMiddleware() (*jwt.GinJWTMiddleware, error) {
    jwtMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
        Realm:      "test zone",
        Key:        []byte("secret key"),
        Timeout:    time.Hour * 24,
        MaxRefresh: time.Hour * 24 * 7,
        SendCookie: false,
        PayloadFunc: func(data interface{}) jwt.MapClaims {
            return jwt.MapClaims{
                jwt.IdentityKey: data,
            }
        },
        Authenticator: func(c *gin.Context) (interface{}, error) {
            var l loginRequest

            if err := c.ShouldBind(&l); err != nil {
                return "", jwt.ErrMissingLoginValues
            }

            if !l.isValid() {
                return "", jwt.ErrFailedAuthentication
            }

            return l.Email, nil
        },
    })

    if err != nil {
        return nil, err
    }

    err = jwtMiddleware.MiddlewareInit()

    if err != nil {
        return nil, err
    }

    return jwtMiddleware, nil
}

type loginRequest struct {
    Email    string `form:"email" json:"email" binding:"required"`
    Password string `form:"password" json:"password" binding:"required"`
}

func (l loginRequest) isValid() bool {
    // TODO : 一般的にはデータベースやストレージ、SaaSから取得する
    passwords := map[string]string{
        "admin@gmail.com": "admin",
        "test@gmail.com":  "test",
    }

    return passwords[l.Email] == l.Password
}
