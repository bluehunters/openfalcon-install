Install_Dashboard()
{
	cd $WORKSPACE
	git clone https://github.com/open-falcon/dashboard.git
	yum install -y python-virtualenv
	yum install -y python-devel
	yum install -y openldap-devel
	yum install -y mysql-devel
	yum groupinstall "Development tools"
	cd $WORKSPACE/dashboard/
	virtualenv ./env
#	./env/bin/pip install -r pip_requirements.txt -i https://pypi.douban.com/simple
	./env/bin/pip install -r pip_requirements.txt
}
