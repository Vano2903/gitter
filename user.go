package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/Vano2903/gitter/internal/email"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientUser                *mongo.Client
	ctxUser                   context.Context
	collectionUser            *mongo.Collection
	collectionUserUnconfirmed *mongo.Collection
)

type User struct {
	ID     primitive.ObjectID `bson:"_id, omitempty" json:"-"`
	User   string             `bson:"user, omitempty" json:"user, omitempty"`          //username
	Email  string             `bson:"email, omitempty"  json:"email, omitempty"`       //email
	Pass   string             `bson:"password, omitempty"  json:"password, omitempty"` //password
	Salt   string             `bson:"salt, omitempty"  json:"-"`
	PfpUrl string             `bson:"pfp_url, omitempty" json:"pfp_url, omitempty"` //url of the profile picture
}

//check if the structure has empty fields
func (x User) IsStructureEmpty() bool {
	return reflect.DeepEqual(x, User{})
}

//will connect to database on user's collection
func ConnectToDatabaseUsers() error {
	ctxUser, _ := context.WithTimeout(context.TODO(), 10*time.Second)

	//try to connect
	clientOptions := options.Client().ApplyURI(conf.Uri)
	clientUser, err := mongo.Connect(ctxUser, clientOptions)
	if err != nil {
		return err
	}

	//check if connection is established
	if err := clientUser.Ping(context.TODO(), nil); err != nil {
		return err
	}

	//assign to the global variable "collection" the users' collection
	collectionUser = clientUser.Database("gitter").Collection("users")
	collectionUserUnconfirmed = clientUser.Database("gitter").Collection("users-unconfirmed")
	return nil
}

//return the url of a user's pfp given the username
func GetProfilePicture(username string) (string, error) {
	query := bson.M{"user": username}
	cur, err := collectionUser.Find(ctxUser, query)
	if err != nil {
		return "", err
	}
	defer cur.Close(ctxUser)
	var userFound []User

	//convert cur in []User
	if err = cur.All(context.TODO(), &userFound); err != nil {
		return "", err
	}
	if len(userFound) > 0 {
		return userFound[0].PfpUrl, nil
	}
	return "", errors.New("no user found as " + username)
}

//return the user based on username and password
func QueryUser(user, pass string) (User, error) {
	//create a query using email and password
	query := bson.M{"user": user, "password": pass}
	//check the db for the credentials using the query
	cur, err := collectionUser.Find(ctxUser, query)
	if err != nil {
		return User{}, err
	}
	defer cur.Close(ctxUser)
	var userFound []User

	//convert cur in []User
	if err = cur.All(context.TODO(), &userFound); err != nil {
		return User{}, err
	}
	if len(userFound) == 0 {
		return User{}, errors.New("no user found as " + user)
	}
	//return the first user found (since using email and password will only return a slice of 1)
	return userFound[0], nil
}

func QueryByEmail(email string) (User, error) {
	//create a query using email and password
	query := bson.M{"email": email}
	//check the db for the credentials using the query
	cur, err := collectionUser.Find(ctxUser, query)
	if err != nil {
		return User{}, err
	}
	defer cur.Close(ctxUser)
	var userFound []User

	//convert cur in []User
	if err = cur.All(context.TODO(), &userFound); err != nil {
		return User{}, err
	}
	if len(userFound) == 0 {
		return User{}, errors.New("no user found as " + email)
	}
	//return the first user found (since using email and password will only return a slice of 1)
	return userFound[0], nil
}

//check that the username, email and password are not empty
//check if the email is valid
//check if the username has invalid characters (since it will be used as the name of the user's repo)
//the length of the username must be between 4 and 20
func CheckUserCreationInfo(user, emailUser, pass string) error {
	if user == "" || pass == "" || emailUser == "" {
		return errors.New("uncorrect/uncomplete credentials to create the user")
	}
	if !email.IsValid(emailUser) {
		return errors.New("email is not valid")
	}
	if len(user) < 4 || len(user) > 20 {
		return errors.New("username must be longer than 4 characters")
	}
	if strings.Contains(user, "/") {
		return errors.New("username can't contain '/' character")
	}
	if strings.Contains(user, "\\") {
		return errors.New("username can't contain '\\' character")
	}
	if user == "." || user == ".." {
		return errors.New("username can't be '.' or '..'")
	}
	return nil
}

//will add the user to database, return the id if succeded adding the user
func AddUser(user, email, pass, salt string, confirmed bool) (int, error) {
	if err := CheckUserCreationInfo(user, email, pass); err != nil {
		return 406, fmt.Errorf("error with the credentials given: %s", err.Error())
	}

	//check if not already registered
	//if nil is returned means that a user was found
	if _, err := QueryUser(user, pass); err == nil {
		return 400, fmt.Errorf("User already exist: %s", err.Error())
	}

	if salt != "" {
		//generating a new randomizer
		var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
		//generating the user salt (random string of length 16)
		var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		b := make([]byte, 16)
		for i := range b {
			b[i] = letters[seededRand.Intn(len(letters))]
		}
		salt = string(b)
	}
	//adding a default profile picture
	pfpUrl := "https://avatars.dicebear.com/api/identicon/" + user + ".svg"
	//adding user to database
	//This struct is needed cause user has ID field
	toInsert := struct {
		User   string `bson:"user, omitempty"      json: "user, omitempty"`
		Email  string `bson:"email, omitempty"  json:"email, omitempty"`
		Salt   string `bson:"salt, omitempty"  json: "-"`
		Pass   string `bson:"password, omitempty"  json: "password, omitempty"`
		PfpUrl string `bson:"pfp_url, omitempty" json:"pfp_url, omitempty"`
	}{
		user,
		email,
		salt,
		fmt.Sprintf("%x", sha256.Sum256([]byte(pass+":"+salt))),
		pfpUrl,
	}

	if confirmed {
		//add the user to the database
		if _, err := collectionUser.InsertOne(ctxUser, toInsert); err != nil {
			return 500, fmt.Errorf("error adding the user to the database: %s", err.Error())
		}
	} else {
		//add the user to the database
		if _, err := collectionUserUnconfirmed.InsertOne(ctxUser, toInsert); err != nil {
			return 500, fmt.Errorf("error adding the user to the database: %s", err.Error())
		}
	}

	//create the repository for the user (no need to run git init)
	if err := CreateNewDir(conf.Repos+user, false); err != nil {
		return 500, fmt.Errorf("error creating the repo: %s", err.Error())
	}

	return 201, nil
}

//delete a user from the database
func DeleteUser(user, pass string, isConfirmed bool) error {
	if isConfirmed {
		_, err := collectionUser.DeleteOne(ctxUser, bson.M{"user": user, "password": pass})
		return err
	}
	_, err := collectionUserUnconfirmed.DeleteOne(ctxUser, bson.M{"user": user, "password": pass})
	return err
}
