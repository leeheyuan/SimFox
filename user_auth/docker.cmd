sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://registry.docker-cn.com"
  ]
} 
EOF
sudo systemctl daemon-reload
sudo systemctl stop docker.service
sudo systemctl stop docker.socket
sudo systemctl start docker.service
sudo systemctl start docker.socket


sudo docker pull docker.mirrors.ustc.edu.cn/ubuntu:latest


export http_proxy=http://192.168.0.105:7897
export https_proxy=http://192.168.0.105:7897
export all_proxy=socks5://192.168.0.105:7897


[Service]
Environment="HTTP_PROXY=http://192.168.0.105:7897/"
Environment="HTTPS_PROXY=http://192.168.0.105:7897/"
Environment="NO_PROXY=localhost,127.0.0.1"



sudo docker run -d \
  --name mysql-server \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=123456 \
  -e MYSQL_DATABASE=simulationtraffic \
  -e MYSQL_USER=simulation \
  -e MYSQL_PASSWORD=simulationSaas \
  -e MYSQL_ALLOW_EMPTY_PASSWORD=no \
  mysql:latest

-- 1）创建用户（'%' 表示任意主机都可连）
CREATE USER 'gogs'@'%' IDENTIFIED BY 'AppPass!23';

-- 2）授权（这里示例授予对所有库的所有权限；生产环境请按需最小化权限）
GRANT ALL PRIVILEGES ON *.* 
  TO 'gogs'@'%' 
  WITH GRANT OPTION;

-- 3）刷新权限表
FLUSH PRIVILEGES;



sudo tee /etc/systemd/system/gogs.service > /dev/null << 'EOF'
[Unit]
Description=Gogs liheyuan Service
After=network.target

[Service]
Type=simple
User=liheyuan
Group=liheyuan
WorkingDirectory=/opt/gogs/gogs
ExecStart=/usr/bin/sudo -u liheyuan /opt/gogs/gogs/gogs web
Restart=always
Environment=USER=liheyuan HOME=/opt/gogs/gogs

[Install]
WantedBy=multi-user.target
EOF
