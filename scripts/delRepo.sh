#!/bin/sh
#$1 is the user name
#$2 is the password
#$3 is the repo to delete

#htpasswd path
PASS="/var/www/html/git/htpasswd"
#name of the repo
REPO="/var/www/html/git/$1/$3.git"
#check the credentials
htpasswd -vb $PASS $1 $2

#0 is success 3 is failure
if [ $? = "0" ]
then
    #check if the repo exists
    if [ -d $REPO ]
    then
        #remove the repo
        rm -r $REPO
        echo "repo deleted successfully"
        exit 0
    else
        echo "repo not existing"
        exit 2
    fi
else
    echo "incorrect credentials"
    exit 1
fi