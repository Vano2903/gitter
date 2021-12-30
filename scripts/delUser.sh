#!/bin/sh
#$1 is the user name
#$2 is the password

#htpasswd path
PASS="/var/www/html/git/htpasswd"
#check the credentials
htpasswd -vb $PASS $1 $2

#0 is success 3 is failure
if [ $? = "0" ]
then
    #remove the user
    htpasswd -D $PASS $1
    STATUS=$?
    #remove the directory
    rm -r "/var/www/html/git/$1"
    exit $STATUS
else
    echo "invalid credentials"
    exit 1
fi