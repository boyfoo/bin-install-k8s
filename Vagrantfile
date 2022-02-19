

Vagrant.configure("2") do |config|

  config.vm.box = "centos7docker19"


  config.vm.define "node01" do |node01|
    node01.vm.hostname = "node01"
    node01.vm.network "private_network", ip: "192.168.33.10"
  end

  config.vm.define "node02" do |node02|
    node02.vm.hostname = "node02"
    node02.vm.network "private_network", ip: "192.168.33.11"
  end

  config.vm.provider "virtualbox" do |vb|
    vb.cpus = 2
    vb.memory = "4096"
  end

end
