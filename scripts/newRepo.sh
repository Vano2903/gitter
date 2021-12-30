#!/bin/sh
#$1 is the user name
#$2 is the password
#$3 is the repo to delete
#htpasswd path
PASS="/var/www/html/git/htpasswd"
#path of the repo to create
REPO="/var/www/html/git/$1/$3.git"

#check the credentials
htpasswd -vb $PASS $1 $2

#0 is success 3 is failure
if [ $? = "0" ]
then
    #check if the repo exists
    if [ -d $REPO ]
    then
        echo "repo already existing"
        exit 2
    else
        #create the directory
        mkdir $REPO
        #set the pwd
        cd $REPO
        :
        #initialize the repo
        git --bare init
        #update the server information
        git update-server-info
        #set the permissions of the folder (nginx)
        chown -R www-data:www-data .
        #set user to read/write/execute and group/global read/execute, octal notation
        chmod -R 755 .
        echo "repo created correctly"
        exit 0
    fi
else
    echo "incorrect credentials"
    exit 1
fi