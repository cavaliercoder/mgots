# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<SCRIPT
#
# Install MongoDB repo
# See: http://docs.mongodb.org/manual/tutorial/install-mongodb-on-ubuntu/
#
apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10
echo "deb http://repo.mongodb.org/apt/ubuntu "$(lsb_release -sc)"/mongodb-org/3.0 multiverse" | tee /etc/apt/sources.list.d/mongodb-org-3.0.list

#
# Install system packages
#
apt-get -y update
apt-get -y install curl git mercurial mongodb-org-server mongodb-org-shell

#
# Install Go
#
# Use upstream tarball instead of Ubuntu package as the Ubuntu version throws a
# "crypto: requested hash function is unavailable" panic whenever attempting to
# install a package from gokpg.in.
#
GOVER=1.4.2
GOPKG=go${GOVER}.linux-amd64.tar.gz
GOURL=https://storage.googleapis.com/golang/${GOPKG}
[[ -f /vagrant/${GOPKG} ]] || curl -sLo /vagrant/${GOPKG} ${GOURL}
if [[ ! -d /usr/local/go ]]; then
	tar -C /usr/local -xzf /vagrant/${GOPKG}
fi

cat > /etc/profile.d/golang.sh <<EOL
export GOPATH=\\${HOME}/gocode
export PATH=/usr/local/go/bin:\\${GOPATH}/bin:\\${PATH}

EOL

#
# Setup Vagrant GOPATH
#
su -m vagrant <<CMD
. /etc/profile.d/golang.sh
export GOPATH=/home/vagrant/gocode
mkdir -p /home/vagrant/gocode/{src,pkg,bin}
mkdir -p /home/vagrant/gocode/src/github.com/cavaliercoder
ln -s /vagrant /home/vagrant/gocode/src/github.com/cavaliercoder/mgots

#
# Install package dependencies for vagrant user
#
go get -v gopkg.in/mgo.v2
go get -v gopkg.in/mgo.v2/bson

CMD

SCRIPT

Vagrant.configure(2) do |config|
  config.vm.box = "hashicorp/precise64"
  config.vm.provision "shell", inline: $script

  config.vm.network "forwarded_port", guest: 6060, host: 6060 # go doc
end
