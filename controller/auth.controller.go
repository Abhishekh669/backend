package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"github.com/Abhishekh669/backend/services"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GoogleUserType struct {
	ID            string `json:"id,omitempty" bson:"id,omitempty"`
	Email         string `json:"email,omitempty" bson:"email,omitempty"`
	VerifiedEmail bool   `json:"verified_email" bson:"verified_email,omitempty"`
	Name          string `json:"name,omitempty" bson:"name,omitempty"`
	GivenName     string `json:"given_name,omitempty" bson:"given_name,omitempty"`
	Picture       string `json:"picture,omitempty" bson:"picture,omitempty"`
}

func generateState() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)

	if err != nil {
		return "", fmt.Errorf("failed to generate state : %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// func generateJwtToken(user model.User) (string, error) {
// 	jwtSecret := []byte(os.Getenv("SESSION_SECRET"))

// 	claims := jwt.MapClaims{
// 		"sub": user.ID,
// 		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
// 		"iat": time.Now().Unix(),
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	signedToken, err := token.SignedString(jwtSecret)

// 	if err != nil {
// 		return "", err
// 	}

// 	return signedToken, nil

// }

var server_state string

func AuthHandler(w http.ResponseWriter, r *http.Request) {

	state, err := generateState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	server_state = state

	fmt.Println("this ishte state in the login  : ", state)

	url := configuration.GoogleAuthConfig.AuthCodeURL(server_state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func AuthGoogleCallback(w http.ResponseWriter, r *http.Request) {

	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	fmt.Println("this is hte state  in callback : ", r.FormValue("state"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var userInfo GoogleUserType

	err = json.Unmarshal(content, &userInfo)

	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//check the user now

	existingUser, err := services.CheckUser(userInfo.Email)
	fmt.Println("HTisi s erro in lofin : ", err)

	var newUser model.User

	if err == nil {
		//user exists
		newUser = existingUser

	} else {

		//create new user
		userData := model.User{
			Name:        userInfo.Name,
			Email:       userInfo.Email,
			Image:       userInfo.Picture,
			IsAdmin:     false,
			IsOnBoarded: false,
			AllowCreate: false,
			IsVerified:  false,
		}

		createdUser, err := services.CreateUser(userData)

		if err != nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		newUser = createdUser

	}

	var (
		sessionKey      = []byte(os.Getenv("SESSION_SECRET"))
		sessionStore    = sessions.NewCookieStore(sessionKey)
		sessionName     = "__session"
		sessionLifeTime = 7 * 24 * time.Hour // Session lasts for 7 days
	)

	//create session

	session, _ := sessionStore.Get(r, sessionName)

	session.Values["user_id"] = newUser.ID.Hex()
	session.Values["email"] = newUser.Email
	session.Values["name"] = newUser.Name
	session.Values["authenticated"] = true

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(sessionLifeTime.Seconds()),
		HttpOnly: true,
		Secure:   false, //TODO : true in production only
		SameSite: http.SameSiteNoneMode,
	}

	err = session.Save(r, w)

	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}
	responseData := map[string]interface{}{
		"message": "user authenticated successfully",
		"user":    newUser,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)

}

func getUserInfo(state string, code string) ([]byte, error) {

	if state == "" {
		return nil, fmt.Errorf("invalid auth state")
	}
	token, err := configuration.GoogleAuthConfig.Exchange(context.Background(), code)

	if err != nil {
		return nil, fmt.Errorf("code exchange failed : %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return contents, nil

}

func GetValidateSession(w http.ResponseWriter, r *http.Request) {

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	// Check if the user is authenticated in the session
	if !sessionData.Authenticated {
		http.Error(w, "User not authenticated", http.StatusInternalServerError)
		return
	}

	// Retrieve the user ID from the session
	if sessionData.UserId == "" {
		http.Error(w, "User not authenticated", http.StatusInternalServerError)
		return
	}

	ObjectUserId, err := primitive.ObjectIDFromHex(sessionData.UserId)

	if err != nil {
		http.Error(w, "Failed to parse userid ", http.StatusInternalServerError)
		return
	}

	// Retrieve the user from the database using the user ID
	user, err := services.GetUserById(ObjectUserId)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Return the authenticated user data
	responseData := map[string]interface{}{
		"message": "user authenticated successfully",
		"user": map[string]interface{}{
			"address":       user.Address,
			"age":           user.Age,
			"isAdmin":       user.IsAdmin,
			"codeName":      user.CodeName,
			"createdAt":     user.CreatedAt,
			"email":         user.Email,
			"field":         user.Field,
			"mainField":     user.MainField,
			"isOnBoarded":   user.IsOnBoarded,
			"isVerified":    user.IsVerified,
			"name":          user.Name,
			"phoneNumber":   user.PhoneNumber,
			"_id":           user.ID,
			"qualification": user.Qualification,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)

}
