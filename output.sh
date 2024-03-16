#!/bin/bash


help="Usage: issue ID \"TITLE\" BRANCH"
ISSUE=${1:?$help}
TITLE=${2:?$help}
BRANCH=${3:?$help}

echo ISSUE: $ISSUE
echo TITLE: $TITLE
echo BRANCH: $BRANCH


