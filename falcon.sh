#export FALCON_HOME=/home/work
#export WORKSPACE=$FALCON_HOME/open-falcon
#mkdir -p $WORKSPACE

Install_Openfalcon()
{	
	cd $curdir/src
	if [ -s open-falcon-v$OF_VERSION.tar.gz ];then
	tar -zxvf open-falcon-v$OF_VERSION.tar.gz -C $WORKSPACE
	else
	wget https://github.com/open-falcon/falcon-plus/releases/download/$OF_VERSION/open-falcon-v$OF_VERSION.tar.gz
	tar -zxvf open-falcon-v$OF_VERSION.tar.gz -C $WORKSPACE
	fi
}
Start_Openfalcon()
{
	cd $WORKSPACE
	./open-falcon start
	./open-falcon check
}

Start_Single()
{	
	echo "1. agent"
	echo "2. graph"


}

Init_Conf()
{	
	cd $WORKSPACE
	echo $ROOT_PASSWORD
	sed -i "s#root:@tcp#root:$ROOT_PASSWORD@tcp#g"  ./api/config/cfg.json
	sed -i "s#root:@tcp#root:$ROOT_PASSWORD@tcp#g"  ./nodata/config/cfg.json
	sed -i "s#root:@tcp#root:$ROOT_PASSWORD@tcp#g"	./alarm/config/cfg.json
	sed -i "s#root:@tcp#root:$ROOT_PASSWORD@tcp#g"	./hbs/config/cfg.json
	sed -i "s#root:@tcp#root:$ROOT_PASSWORD@tcp#g"	./aggregator/config/cfg.json
	sed -i "s#root:@tcp#root:$ROOT_PASSWORD@tcp#g"	./graph/config/cfg.json
	sed -i "s#\"\"#\"$ROOT_PASSWORD\"#g"  ./dashboard/rrd/config.py
}
