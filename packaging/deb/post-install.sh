#!/usr/bin/env bash

NAME=lora-app-server
BIN_DIR=/usr/bin
SCRIPT_DIR=/usr/lib/$NAME/scripts
LOG_DIR=/var/log/$NAME
DAEMON_USER=appserver
DAEMON_GROUP=appserver

function install_init {
	cp -f $SCRIPT_DIR/init.sh /etc/init.d/$NAME
	chmod +x /etc/init.d/$NAME
	update-rc.d $NAME defaults
}

function install_systemd {
	cp -f $SCRIPT_DIR/$NAME.service /lib/systemd/system/$NAME.service
	systemctl daemon-reload
	systemctl enable $NAME
}

function restart_appserver {
	echo "Restarting LoRa App Server"
	which systemctl &>/dev/null
	if [[ $? -eq 0 ]]; then
		systemctl daemon-reload
		systemctl restart $NAME
	else
		/etc/init.d/$NAME restart || true
	fi	
}

# create appserver user
id $DAEMON_USER &>/dev/null
if [[ $? -ne 0 ]]; then
	useradd --system -U -M $DAEMON_USER -d /bin/false
fi

mkdir -p "$LOG_DIR"
chown $DAEMON_USER:$DAEMON_GROUP "$LOG_DIR"

# create configuration directory
if [[ ! -d /etc/$NAME ]]; then
	mkdir /etc/$NAME
	chown $DAEMON_USER:$DAEMON_GROUP /etc/$NAME
	chmod 750 /etc/$NAME
fi

# migrate old environment variable based configuration to new format and
# path.
if [[ -f /etc/default/$NAME && ! -f /etc/$NAME/$NAME.toml ]]; then
	set -a
	source /etc/default/$NAME
	lora-app-server configfile > /etc/$NAME/$NAME.toml
	chown $DAEMON_USER:$DAEMON_GROUP /etc/$NAME/$NAME.toml
	chmod 640 /etc/$NAME/$NAME.toml
	mv /etc/default/$NAME /etc/default/$NAME.backup

	echo -e "\n\n\n"
	echo "-----------------------------------------------------------------------------------------"
	echo "Your configuration file has been migrated to a new location and format!"
	echo "Path: /etc/$NAME/$NAME.toml"
	echo "-----------------------------------------------------------------------------------------"
	echo -e "\n\n\n"
fi

# create example configuration file
if [[ ! -f /etc/$NAME/$NAME.toml ]]; then
	HTTP_TLS_CERT=/etc/$NAME/certs/http.pem HTTP_TLS_KEY=/etc/$NAME/certs/http-key.pem lora-app-server configfile > /etc/$NAME/$NAME.toml
	chown $DAEMON_USER:$DAEMON_GROUP /etc/$NAME/$NAME.toml
	chmod 640 /etc/$NAME/$NAME.toml
	echo -e "\n\n\n"
	echo "-------------------------------------------------------------------------------------"
	echo "A sample configuration file has been copied to: /etc/$NAME/$NAME.toml"
	echo "After setting the correct values, run the following command to start LoRa App Server:"
	echo ""
	which systemctl &>/dev/null
	if [[ $? -eq 0 ]]; then
		echo "$ sudo systemctl start $NAME"
	else
		echo "$ sudo /etc/init.d/$NAME start"
	fi
	echo "-------------------------------------------------------------------------------------"
	echo -e "\n\n\n"
fi

# create self-signed certificate if not certificate file exists
if [[ ! -r /etc/$NAME/certs/http-key.pem ]] && [[ ! -r /etc/$NAME/certs/http.pem ]]; then
	mkdir -p /etc/$NAME/certs
	openssl req -x509 -newkey rsa:4096 -keyout /etc/$NAME/certs/http-key.pem -out /etc/$NAME/certs/http.pem -days 365 -nodes -batch -subj "/CN=localhost"
	chown -R $DAEMON_USER:$DAEMON_GROUP /etc/$NAME
	chmod 640 /etc/$NAME/certs/*.pem
	echo -e "\n\n\n"
	echo "-------------------------------------------------------------------------------------------"
	echo "A self-signed TLS certificate has been generated and written to: /etc/$NAME/certs"
	echo ""
	echo "This is convenient for testing, but should be replaced with a proper certificate!" 
	echo "-------------------------------------------------------------------------------------------"
	echo -e "\n\n\n"
fi

# add start script
which systemctl &>/dev/null
if [[ $? -eq 0 ]]; then
	install_systemd
else
	install_init
fi

if [[ -n $2 ]]; then
	restart_appserver
fi

