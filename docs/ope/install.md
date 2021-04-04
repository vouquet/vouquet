Vouquetのサーバインストール手順
===

## 1. mysqlインストールや準備

1. install
	```
	sudo apt install mysql-server
	```
2. secure installation
	```
	<space>mysql_secure_installation -p{new-password} -D ##先頭spaceでhistroy除外
	```
3. create databases
	```
	mysql -u root -p
	$ DROP DATABASE IF EXISTS vouquet;
	$ CREATE DATABASE IF NOT EXISTS vouquet CHARACTER SET utf8 COLLATE utf8_general_ci;
	$ CREATE USER 'vouquet'@'localhost' identified by '<your password>;
	$ GRANT ALL PRIVILEGES ON vouquet.* to 'vouquet'@'localhost';
	$ GRANT RELOAD ON *.* TO 'vouquet'@'localhost';
	$ FLUSH PRIVILEGES;
	```

## 2. 動作アカウントの作成

1. `sudo useradd vouquet --system`

## 3. 動作ファイル設置
1. ディレクトリ作成
	1. `sudo mkdir -p /var/service/vouquet`
	1. `sudo mkdir -p /var/service/vouquet/bin`
	1. `sudo mkdir -p /var/service/vouquet/sbin`
	1. `sudo mkdir -p /var/service/vouquet/etc`
2. 実行ファイル設置
	1. `sudo cp <vouquet repository>/src/go/bin/* /var/service/vouquet/bin/`
	1. `sudo cp <vouquet repository>/src/go/bin/vqt_eval /var/service/vouquet/sbin/`
	1. `sudo cp <vouquet repository>/etc.in/systemd/system/* /etc/systemd/system/`
		* `systemd` に、実行用のパラメータが書かれているので、適宜修正する
3. 権限変更
	1. `sudo chown -R vouquet:vouquet /var/service/vouquet`
4. 設定ファイル設置
	1. `sudo vim /var/service/vouquet/etc/vouquet.conf`
		* `<vouquet repository>/sample/vouquet.conf`を参考
		* `DBPass`には、mysql設定時に指定した、`{new-password}`を指定する
	2. `sudo chown vouquet:vouquet /var/service/vouquet/etc/vouquet.conf`
	3. `sudo chmod 600 /var/service/vouquet/etc/vouquet.conf`

## 4. 起動と有効化
1. `sudo systemctl daemon-reload`
2. `sudo systemctl enable vqt_registrar vqt_noticer vqt_florister`
3. `sudo systemctl restart vqt_registrar vqt_noticer vqt_florister`
