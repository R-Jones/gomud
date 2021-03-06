curl -sL https://deb.nodesource.com/setup_9.x -o nodesource_setup.sh;
sudo bash nodesource_setup.sh;
sudo apt-get -y install postgresql postgresql-contrib libpq-dev nodejs; 
sudo su - postgres -c "psql -c \"CREATE USER gomud WITH PASSWORD 'gomud'\"";
sudo su - postgres -c "psql -c \"ALTER ROLE gomud SET client_encoding TO 'utf8'\"";
sudo su - postgres -c "psql -c \"ALTER ROLE gomud SET default_transaction_isolation TO 'read committed'\"";
sudo su - postgres -c "psql -c \"ALTER ROLE gomud SET timezone TO 'utc'\"";
sudo su - postgres -c "psql -c \"CREATE DATABASE gomud\"";
sudo su - postgres -c "psql -c \"GRANT ALL PRIVILEGES ON DATABASE gomud TO gomud\"";
curl -O https://storage.googleapis.com/golang/go1.10.linux-amd64.tar.gz;
tar xvf go1.10.linux-amd64.tar.gz; 
sudo chown -R root:root ./go;
sudo mv go /usr/local;
sudo echo "export GOPATH=$HOME/work" >> ~/.profile;
sudo echo "export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin" >> ~/.profile;
. ~/.profile;
go get -u github.com/go-pg/pg;
go get -u github.com/gorilla/websocket;
go get golang.org/x/crypto/bcrypt;
cd frontend;
npm install;
