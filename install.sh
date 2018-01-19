curdir=`pwd`
echo "Current dir is \"$curdir\""
export FALCON_HOME=/home/work
export WORKSPACE=$FALCON_HOME/open-falcon
mkdir -p $WORKSPACE

. config.sh
. mysql.sh
. redis.sh
. golang.sh
. falcon.sh
. dashboard.sh

echo "Start install Openfalcon"
Install_Redis
Install_Mysql
Install_Golang
Install_Openfalcon
Install_Dashboard
Init_Conf
Start_Openfalcon
cd $WORKSPACE/dashboard && ./control start

