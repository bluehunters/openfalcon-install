Install_Golang()
{
	echo "We are installing golang now!"
	cd $curdir/src
	if [ -s go1.9.2.linux-amd64.tar.gz  ] ;then
	rm -rf ./go
	tar zxvf go$GO_VERSION.linux-amd64.tar.gz 
	echo go$GO_VERSION
#	sleep 10
	else 
	wget https://studygolang.com/dl/golang/go1.9.2.linux-amd64.tar.gz -O ./src/
	echo "Download Successful."
	tar zxvf go$GO_VERSION.linux-amd64.tar.gz
	fi
	cd go && mkdir /usr/local/go
	cp -rf ./bin/ /usr/local/go
	ln -s /usr/local/go/bin/go /usr/bin/go
	export PATH=$PATH:/usr/local/go/bin
	
	

}

