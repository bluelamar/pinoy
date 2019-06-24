
echo "Rotating..."
#NAME=`mktemp -u pinoy.log_XXX`
SUFFIX=`date --iso-8601=ns`
NEW_NAME="pinoy.log_$SUFFIX"
echo "move log to name=$NEW_NAME"
mv /var/log/pinoy/pinoy.log /var/log/pinoy/$NEW_NAME
killall -1 pinoy

