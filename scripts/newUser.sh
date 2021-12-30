#!/bin/sh
#$1 is the user name
#$2 is the password

#user repo path
USERREPO="/var/www/html/git/$1"

#check if the repo exist
if [ -d $USERREPO ]
then
    echo "user already created"
    exit 1
else
    #create the user
    htpasswd -cbB "/var/www/html/git/htpasswd" $1 $2
    #create the repo
    mkdir $USERREPO
    #set the permissions of the folder (nginx)
    chown -R www-data:www-data .
    #set user to read/write/execute and group/global read/execute, octal notation
    chmod -R 755 .
    echo "user added correctly"
    exit 0
fi