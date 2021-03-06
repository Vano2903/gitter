package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
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

type Commit struct {
	Tree    string `json:"tree, omitempty"`
	Hash    string `json:"hash, omitempty"`
	Message string `json:"message, omitempty"`
	Date    string `json:"date, omitempty"`
}

type Object struct {
	Type string `json:"type"`
	Hash string `json:"hash"`
	Name string `json:"name"`
}

type Branch struct {
	Name string `json: "name"`
	Hash string `json: "hash"`
	Type string `json: "type"`
}

type Info struct {
	CommitsNum int      `json:"commits_num"`
	Commits    []Commit `json:"commits"`
	Files      []Object `json:"files"`
	LastCommit Commit   `json:"last_commit"`
	Branches   []Branch `json:"branches"`
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
func QueryUser(user, pass string, onUnconfirmed bool) (User, error) {
	//create a query using email and password
	query := bson.M{"user": user, "password": pass}
	//check the db for the credentials using the query
	var cur *mongo.Cursor
	var err error
	if onUnconfirmed {
		cur, err = collectionUserUnconfirmed.Find(ctxUser, query)
	} else {
		cur, err = collectionUser.Find(ctxUser, query)
	}
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

func QueryByEmail(email string, inRegistered bool) (User, error) {
	//create a query using email and password
	query := bson.M{"email": email}
	//check the db for the credentials using the query
	var cur *mongo.Cursor
	var err error
	if inRegistered {
		cur, err = collectionUser.Find(ctxUser, query)
	} else {
		cur, err = collectionUserUnconfirmed.Find(ctxUser, query)
	}
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

func QueryByUsername(username string, inRegistered bool) (User, error) {
	//create a query using username and password
	query := bson.M{"user": username}
	//check the db for the credentials using the query
	var cur *mongo.Cursor
	var err error
	if inRegistered {
		cur, err = collectionUser.Find(ctxUser, query)
	} else {
		cur, err = collectionUserUnconfirmed.Find(ctxUser, query)
	}
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
		return User{}, errors.New("no user found as " + username)
	}
	//return the first user found (since using username and password will only return a slice of 1)
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
		return errors.New("username must be between 4 and 20 characters")
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
	if _, err := QueryUser(user, pass, false); err == nil {
		return 400, fmt.Errorf("User already exist: %s", err.Error())
	}

	if _, err := QueryUser(user, pass, true); err == nil {
		return 400, fmt.Errorf("User already exist: %s", err.Error())
	}

	//if salt is null it means that the user is confirmating the account so the salt has already been generated
	if salt == "" {
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
	if !confirmed {
		pass = fmt.Sprintf("%x", sha256.Sum256([]byte(pass+":"+salt)))
	}
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
		pass,
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
	if confirmed {
		if err := CreateNewDir(conf.Repos+user, false); err != nil {
			return 500, fmt.Errorf("error creating the repo: %s", err.Error())
		}
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

//create a new repo and run git init --bare
func (u User) CreateRepo(repoName string) error {
	if err := CreateNewDir(conf.Repos+u.User+"/"+repoName+".git", true); err != nil {
		return fmt.Errorf("error creating the repo: %s", err.Error())
	}
	return nil
}

//return a list with all the repos of a user
func (u User) GetRepos() ([]string, error) {
	files, err := ioutil.ReadDir(conf.Repos + u.User)
	if err != nil {
		return nil, err
	}

	var repos []string

	for _, f := range files {
		repos = append(repos, f.Name()[:len(f.Name())-4])
	}
	return repos, nil
}

//get the info of a repo from the filesystem
func (u User) GetRepoInfo(repo string) (Info, error) {
	var info Info
	//get the branches
	cmd := exec.Command("git", "for-each-ref")
	cmd.Dir = conf.Repos + u.User + "/" + repo + ".git"
	out, err := cmd.Output()
	fmt.Println("for-each:", string(out))
	fmt.Println(err)
	if err != nil {
		return info, err
	}
	branches := strings.Split(string(out), "\n")[:len(strings.Split(string(out), "\n"))-1]
	fmt.Println("branches: ", branches, "len: ", len(branches))
	for _, b := range branches {
		var branch Branch
		if strings.Contains(b, "refs/heads/") {
			branch.Hash = strings.Split(b, " ")[0]
			branch.Type = strings.Split(strings.Split(b, " ")[1], "\t")[0]
			branch.Name = strings.Replace(strings.Split(b, "\t")[1], "refs/heads/", "", 1)
			info.Branches = append(info.Branches, branch)
		}
	}

	//get all the commits, the number of them and the last commit
	cmd = exec.Command("git", "log", `--pretty=format:%h %ct %s`)
	cmd.Dir = conf.Repos + u.User + "/" + repo + ".git"
	out, _ = cmd.Output()
	fmt.Println("log:", string(out))
	// fmt.Println(err)
	// if err != nil {
	// 	return info, err
	// }
	commits := strings.Split(string(out), "\n")[:len(strings.Split(string(out), "\n"))-1]
	fmt.Println("commits: ", branches, "len: ", len(branches))

	info.CommitsNum = len(commits)
	for i, c := range commits {
		var commit Commit
		commit.Hash = strings.Split(c, " ")[0]
		commit.Date = strings.Split(c, " ")[1]
		commit.Message = strings.Join(strings.Split(c, " ")[2:], " ")
		info.Commits = append(info.Commits, commit)
		if i == 0 {
			info.LastCommit = commit
		}
	}

	//get all the files
	cmd = exec.Command("git", "ls-tree", "-r", "HEAD")
	cmd.Dir = conf.Repos + u.User + "/" + repo + ".git"
	out, _ = cmd.Output()
	fmt.Println("ls-tree:", string(out))
	// fmt.Println(err)
	// if err != nil {
	// 	return info, err
	// }
	files := strings.Split(string(out), "\n")[:len(strings.Split(string(out), "\n"))-1]
	fmt.Println("files: ", branches, "len: ", len(branches))

	for _, f := range files {
		var file Object
		file.Type = strings.Split(f, " ")[1]
		file.Hash = strings.Split(strings.Split(f, " ")[2], "\t")[0]
		file.Name = strings.Split(strings.Split(f, " ")[2], "\t")[1]
		info.Files = append(info.Files, file)
	}

	return info, nil
}

//get the type of hash (commit/tree/blob)
func (u User) GetHashType(repo string, hash string) (string, error) {
	cmd := exec.Command("git", "cat-file", "-t", hash)
	cmd.Dir = conf.Repos + u.User + "/" + repo + ".git"
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Split(string(out), "\n")[0], nil
}

//return a commit with commit hash empty but the tree hash
func (u User) GetCommitInfo(repo string, hash string) (Commit, error) {
	var commit Commit
	//get the commit message
	cmd := exec.Command("git", "show", hash, "--pretty=format:%t %ct %s", "--no-patch")
	cmd.Dir = conf.Repos + u.User + "/" + repo + ".git"
	out, err := cmd.Output()
	if err != nil {
		return commit, err
	}
	commit.Tree = strings.Split(string(out), " ")[0]
	commit.Date = strings.Split(string(out), " ")[1]
	commit.Message = strings.Join(strings.Split(string(out), " ")[2:], " ")

	return commit, nil
}

//return a list of object (files) with the tree hash
func (u User) GetTreeInfo(repo string, hash string) ([]Object, error) {
	var objects []Object
	cmd := exec.Command("git", "ls-tree", "-r", hash)
	cmd.Dir = conf.Repos + u.User + "/" + repo + ".git"
	out, err := cmd.Output()
	if err != nil {
		return objects, err
	}
	files := strings.Split(string(out), "\n")[:len(strings.Split(string(out), "\n"))-1]
	for _, f := range files {
		var file Object
		file.Type = strings.Split(f, " ")[1]
		file.Hash = strings.Split(strings.Split(f, " ")[2], "\t")[0]
		file.Name = strings.Split(strings.Split(f, " ")[2], "\t")[1]
		objects = append(objects, file)
	}
	return objects, nil
}

func (u User) GetBlobInfo(repo string, hash string) (string, error) {
	cmd := exec.Command("git", "cat-file", "-p", hash)
	cmd.Dir = conf.Repos + u.User + "/" + repo + ".git"
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

//check if a directory exists
func (u User) ExistDir(repo string) (bool, error) {
	_, err := os.Stat(conf.Repos + u.User + "/" + repo + ".git")
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
