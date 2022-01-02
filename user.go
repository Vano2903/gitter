package main

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientUser     *mongo.Client
	ctxUser        context.Context
	collectionUser *mongo.Collection
)

type User struct {
	ID     primitive.ObjectID `bson:"_id, omitempty" json:"-"`
	User   string             `bson:"user, omitempty" json:"user, omitempty"`          //username
	Email  string             `bson:"email, omitempty"  json:"email, omitempty"`       //email
	Pass   string             `bson:"password, omitempty"  json:"password, omitempty"` //password
	PfpUrl string             `bson:"pfp_url, omitempty" json:"pfp_url, omitempty"`    //url of the profile picture
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
	//return the first user found (since using email and password will only return a slice of 1)
	return userFound[0], nil
}

//check that the username, email and password are not empty
//check if the username has invalid characters (since it will be used as the name of the user's repo)
//the length of the username must be between 4 and 20
func CheckUserCreationInfo(user, email, pass string) error {
	//check if strings are empty and authlvl between 0 and 2
	if user == "" || pass == "" || email == "" {
		return errors.New("uncorrect/uncomplete credentials to create the user")
	}
	if len(user) <= 4 || len(user) >= 20 {
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
//type 0 = success, 1 = 4xx response, 2 = 5xx response
func AddUser(user, email, pass string) (int, error) {
	if err := CheckUserCreationInfo(user, email, pass); err != nil {
		return 1, err
	}

	//check if not already registered
	//if nil is returned means that a user was found
	if _, err := QueryUser(user, pass); err == nil {
		return 1, errors.New("user already exist")
	}

	//adding a default profile picture
	pfpUrl := "https://avatars.dicebear.com/api/identicon/" + user + ".svg"
	//adding user to database
	//This struct is needed cause user has ID field
	toInsert := struct {
		User   string `bson:"user, omitempty"      json: "user, omitempty"`
		Email  string `bson:"email, omitempty"  json:"email, omitempty"`
		Pass   string `bson:"password, omitempty"  json: "password, omitempty"`
		PfpUrl string `bson:"pfp_url, omitempty" json:"pfp_url, omitempty"`
	}{
		user,
		email,
		pass,
		pfpUrl,
	}

	//add the user to the database
	if _, err := collectionUser.InsertOne(ctxUser, toInsert); err != nil {
		return 2, err
	}

	//create the repository for the user (no need to run git init)
	if err := CreateNewDir(conf.Repos+user, false); err != nil {
		return 2, err
	}

	return 0, nil
}
