Install_Mysql()
{
	#yum install mysql-server
	cd /tmp/ && git clone https://github.com/open-falcon/falcon-plus.git 
	cd /tmp/falcon-plus/scripts/mysql/db_schema/
	mysql -h 127.0.0.1 -u root -p123456 < 1_uic-db-schema.sql
	mysql -h 127.0.0.1 -u root -p123456 < 2_portal-db-schema.sql
	mysql -h 127.0.0.1 -u root -p123456 < 3_dashboard-db-schema.sql
	mysql -h 127.0.0.1 -u root -p123456 < 4_graph-db-schema.sql
	mysql -h 127.0.0.1 -u root -p123456 < 5_alarms-db-schema.sql
#	rm -rf /tmp/falcon-plus/

}
