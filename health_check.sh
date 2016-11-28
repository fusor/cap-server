#!/bin/bash

host=$1
curl -s -o /dev/null -w "%{http_code}" http://$host
