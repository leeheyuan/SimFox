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


export http_proxy=http://192.168.0.102:7897
export https_proxy=http://192.168.0.102:7897
export all_proxy=socks5://192.168.0.102:7897


[Service]
Environment="HTTP_PROXY=http://192.168.0.102:7897/"
Environment="HTTPS_PROXY=http://192.168.0.102:7897/"
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


cd /mnt/d/item/SimFox/config_Sever


sudo docker run -d \
  --name my-postgres \
  -e POSTGRES_USER=simulation \
  -e POSTGRES_PASSWORD=simulationSaas \
  -e POSTGRES_DB=simulationtraffic \
  -p 5432:5432 \
  -v /my/local/pgdata:/var/lib/postgresql/data \
  postgres:16



CREATE USER 'gogs'@'%' IDENTIFIED BY '123456';

GRANT ALL PRIVILEGES ON gogs.* TO 'gogs'@'%';

FLUSH PRIVILEGES;
