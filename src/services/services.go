package services

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"example.com/main/src/initializer"
	"example.com/main/src/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func Get(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// bsonD(filter, doc)
// bsonM(resBody)
func Signup(c *gin.Context) {
	var body *repository.Sign
	var doc bson.D
	var rt string
	var at string

	//get body
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"err": "failed to bind body",
		})
		return
	}

	//encrypt password
	pwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"err": "failed to hash pwd",
		})
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		wg.Done()
		at = makeHashTest(c, TokenType("at"), body.Email)
	}()
	go func() {
		wg.Done()
		rt = makeHashTest(c, TokenType("rt"), body.Email)
	}()
	wg.Wait()

	//make doc
	doc = bson.D{
		{
			Key:   "email",
			Value: body.Email,
		},
		{
			Key:   "password",
			Value: string(pwd),
		},
		{
			Key:   "rt",
			Value: rt,
		},
	}

	//put to db
	if _, err = initializer.Mongos.Collection.InsertOne(context.TODO(), doc); err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"err": "failed to create user",
		})
		return
	}

	//if success 1.return at/rt token 2. save rt token to db / fail, ret err
	c.JSON(http.StatusCreated, gin.H{
		"at": at,
		"rt": rt,
	})
}

func Singin(c *gin.Context) {
	var body *repository.Sign
	var res bson.M
	var rt string
	var at string

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"err": "failed to bind body",
		})
		return
	}

	//find user, and get pwd from db + bind(dcode) to bson.M(res)
	if err := initializer.Mongos.Collection.FindOne(context.TODO(), bson.D{{Key: "email", Value: body.Email}}).Decode(&res); err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"err": "failed to find user",
		})
		return
	}

	//check pwd, than make at/rt and bind to var
	if pwd, ok := res["password"]; ok {
		if err := bcrypt.CompareHashAndPassword([]byte(pwd.(string)), []byte(body.Password)); err != nil {
			c.JSON(http.StatusNotImplemented, gin.H{
				"err": "invalid pwd",
			})
			return
		} else {
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				defer wg.Done()
				at = makeHashTest(c, TokenType("at"), body.Email)
			}()
			go func() {
				defer wg.Done()
				rt = makeHashTest(c, TokenType("rt"), body.Email)
			}()
			wg.Wait()
		}
	} else {
		c.JSON(http.StatusNotImplemented, gin.H{
			"err": "pwd is missing",
		})
		return
	}

	//update rt
	updateRtToken(c, body.Email, rt)

	//return at/rt
	c.JSON(http.StatusCreated, gin.H{
		"at": at,
		"rt": rt,
	})
}

func Signout(c *gin.Context)  {

}

func RefreshToken(c *gin.Context) {

}

func updateRtToken(c *gin.Context, email, rt string) {
	filter := bson.D{{Key: "email", Value: email}}
	updateDoc := bson.D{{Key: "$set", Value: bson.D{{Key: "rt", Value: rt}}}}

	if _, err := initializer.Mongos.Collection.UpdateOne(context.TODO(), filter, updateDoc); err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"err": "failed to update rt",
		})
		return
	}
}

type TokenType string

const (
	at TokenType = TokenType("at")
	rt           = TokenType("rt")
)

func makeHashTest(c *gin.Context, tokenType TokenType, email string) string {
	switch tokenType {
	case at:
		at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": email,
			"exp":   time.Now().Add(time.Hour * 24 * 1).Unix(),
		})
		if atTokenString, err := at.SignedString([]byte(os.Getenv("SECRET"))); err != nil {
			log.Fatal(err)
			c.JSON(http.StatusNotAcceptable, gin.H{
				"err": "failed to create at token",
			})
			return ""
		} else {
			return atTokenString
		}

	case rt:
		rt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": email,
			"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(),
		})

		if rtTokenString, err := rt.SignedString([]byte(os.Getenv("SECRET"))); err != nil {
			c.JSON(http.StatusNotAcceptable, gin.H{
				"err": "failed to create rt token",
			})
			return ""
		} else {
			return rtTokenString
		}
	}
_:
	return ""
}
