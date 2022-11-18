echo "Creating Logs Folder..."

BASE=$PWD

LOGS_FOLDER_NAME=$(date +%s)
mkdir -p $PWD/logs/$LOGS_FOLDER_NAME

LOGS_FOLDER_PATH=$PWD/logs/$LOGS_FOLDER_NAME
echo "Logs for this session: $LOGS_FOLDER_PATH"

touch $LOGS_FOLDER_PATH/core.log
touch $LOGS_FOLDER_PATH/authconfig.log
touch $LOGS_FOLDER_PATH/api.log

cd $BASE/core/ && go run main.go > $LOGS_FOLDER_PATH/core.log &
cd $BASE/authconfig && go run main.go > $LOGS_FOLDER_PATH/authconfig.log &
# go run ./core/main.go > $LOGS_FOLDER_PATH/core.log &
