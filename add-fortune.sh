#!/bin/sh
# double line separated fortunes. some get cut off but oh well.
echo "Populating fortunes.txt, press Ctrl C when you think its big enough."
for i in $(cat fortunes.txt); do 
fortune -o >> fortunes.txt && echo "" >> fortunes.txt;
done