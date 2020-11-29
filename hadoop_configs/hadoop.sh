#!/bin/bash
## Copy ssh to other machines if namenode 
ssh-keygen -t rsa
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-01.cs.illinois.edu
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-02.cs.illinois.edu
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-03.cs.illinois.edu
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-04.cs.illinois.edu
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-05.cs.illinois.edu
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-06.cs.illinois.edu
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-07.cs.illinois.edu
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-08.cs.illinois.edu
ssh-copy-id -i /home/$1/.ssh/id_rsa.pub $1@fa20-cs425-g13-09.cs.illinois.edu

# Run this in /home/netid
# Download and install hadoop
cd ~
wget https://mirrors.gigenet.com/apache/hadoop/common/hadoop-2.9.2/hadoop-2.9.2.tar.gz
tar -xzvf hadoop-2.9.2.tar.gz
sh ./setup-bashrc.sh $1

# Set up hadoop

echo 'export JAVA_HOME=${JAVA_HOME:-"/usr/lib/jvm/java-1.8.0-openjdk-1.8.0.272.b10-1.el7_9.x86_64"}' >> /home/$1/hadoop-2.9.2/etc/hadoop/hadoop-env.sh
echo 'export HADOOP_CONF_DIR=${HADOOP_CONF_DIR:-"/home/$1/hadoop-2.9.2/etc/hadoop"}' >> /home/$1/hadoop-2.9.2/etc/hadoop/hadoop-env.sh

mkdir -p /home/$1/hadoop_store/hdfs/namenode
mkdir -p /home/$1/hadoop_store/hdfs/datanode

## Copy hadoop configs
rm /home/$1/hadoop-2.9.2/etc/hadoop/hdfs-site.xml
cp ./hdfs-site.xml /home/$1/hadoop-2.9.2/etc/hadoop/
rm /home/$1/hadoop-2.9.2/etc/hadoop/yarn-site.xml
cp ./yarn-site.xml /home/$1/hadoop-2.9.2/etc/hadoop/
rm /home/$1/hadoop-2.9.2/etc/hadoop/core-site.xml
cp ./core-site.xml /home/$1/hadoop-2.9.2/etc/hadoop/
rm /home/$1/hadoop-2.9.2/etc/hadoop/mapred-site.xml
cp ./mapred-site.xml /home/$1/hadoop-2.9.2/etc/hadoop/


# Set up ips
echo 'namenode' >> /home/$1/hadoop-2.9.2/etc/hadoop/masters
echo 'namenode' >> /home/$1/hadoop-2.9.2/etc/hadoop/slaves
echo 'datanode1' >> /home/$1/hadoop-2.9.2/etc/hadoop/slaves #02
echo 'datanode2' >> /home/$1/hadoop-2.9.2/etc/hadoop/slaves #03
echo 'datanode3' >> /home/$1/hadoop-2.9.2/etc/hadoop/slaves #04
echo 'datanode4'>> /home/$1/hadoop-2.9.2/etc/hadoop/slaves #05
echo 'datanode5'>> /home/$1/hadoop-2.9.2/etc/hadoop/slaves #06
echo 'datanode6' >> /home/$1/hadoop-2.9.2/etc/hadoop/slaves #07
echo 'datanode7' >> /home/$1/hadoop-2.9.2/etc/hadoop/slaves #08
echo 'datanode8' >> /home/$1/hadoop-2.9.2/etc/hadoop/slaves #09

## Install hadoop on worker nodes

#cd /home/$1/
#ssh $1@fa20-cs425-g13-02.cs.illinois.edu "rm -r /home/$1/hadoop-2.9.2"
#ssh $1@fa20-cs425-g13-03.cs.illinois.edu "rm -r /home/$1/hadoop-2.9.2"
#ssh $1@fa20-cs425-g13-04.cs.illinois.edu "rm -r /home/$1/hadoop-2.9.2"
#ssh $1@fa20-cs425-g13-05.cs.illinois.edu "rm -r /home/$1/hadoop-2.9.2"
#ssh $1@fa20-cs425-g13-06.cs.illinois.edu "rm -r /home/$1/hadoop-2.9.2"
#ssh $1@fa20-cs425-g13-07.cs.illinois.edu "rm -r /home/$1/hadoop-2.9.2"
#ssh $1@fa20-cs425-g13-08.cs.illinois.edu "rm -r /home/$1/hadoop-2.9.2"
#ssh $1@fa20-cs425-g13-09.cs.illinois.edu "rm -r /home/$1/hadoop-2.9.2"
scp -r hadoop-2.9.2 $1@fa20-cs425-g13-02.cs.illinois.edu:/home/$1
scp -r hadoop-2.9.2 $1@fa20-cs425-g13-03.cs.illinois.edu:/home/$1
scp -r hadoop-2.9.2 $1@fa20-cs425-g13-04.cs.illinois.edu:/home/$1
scp -r hadoop-2.9.2 $1@fa20-cs425-g13-05.cs.illinois.edu:/home/$1
scp -r hadoop-2.9.2 $1@fa20-cs425-g13-06.cs.illinois.edu:/home/$1
scp -r hadoop-2.9.2 $1@fa20-cs425-g13-07.cs.illinois.edu:/home/$1
scp -r hadoop-2.9.2 $1@fa20-cs425-g13-08.cs.illinois.edu:/home/$1
scp -r hadoop-2.9.2 $1@fa20-cs425-g13-09.cs.illinois.edu:/home/$1

# Insert bash scripts to workers
cd /home/$1/cs425-mps/hadoop_configs
scp setup-bashrc.sh $1@fa20-cs425-g13-02.cs.illinois.edu:/home/$1
scp setup-bashrc.sh $1@fa20-cs425-g13-03.cs.illinois.edu:/home/$1
scp setup-bashrc.sh $1@fa20-cs425-g13-04.cs.illinois.edu:/home/$1
scp setup-bashrc.sh $1@fa20-cs425-g13-05.cs.illinois.edu:/home/$1
scp setup-bashrc.sh $1@fa20-cs425-g13-06.cs.illinois.edu:/home/$1
scp setup-bashrc.sh $1@fa20-cs425-g13-07.cs.illinois.edu:/home/$1
scp setup-bashrc.sh $1@fa20-cs425-g13-08.cs.illinois.edu:/home/$1
scp setup-bashrc.sh $1@fa20-cs425-g13-09.cs.illinois.edu:/home/$1
ssh $1@fa20-cs425-g13-02.cs.illinois.edu "sh /home/$1/setup-bashrc.sh $1"
ssh $1@fa20-cs425-g13-03.cs.illinois.edu "sh /home/$1/setup-bashrc.sh $1" 
ssh $1@fa20-cs425-g13-04.cs.illinois.edu "sh /home/$1/setup-bashrc.sh $1" 
ssh $1@fa20-cs425-g13-05.cs.illinois.edu "sh /home/$1/setup-bashrc.sh $1" 
ssh $1@fa20-cs425-g13-06.cs.illinois.edu "sh /home/$1/setup-bashrc.sh $1" 
ssh $1@fa20-cs425-g13-07.cs.illinois.edu "sh /home/$1/setup-bashrc.sh $1" 
ssh $1@fa20-cs425-g13-08.cs.illinois.edu "sh /home/$1/setup-bashrc.sh $1" 
ssh $1@fa20-cs425-g13-09.cs.illinois.edu "sh /home/$1/setup-bashrc.sh $1" 

## Make node dir ssh
ssh $1@fa20-cs425-g13-02.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/datanode"
ssh $1@fa20-cs425-g13-02.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/namenode"
ssh $1@fa20-cs425-g13-03.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/datanode"
ssh $1@fa20-cs425-g13-03.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/namenode"
ssh $1@fa20-cs425-g13-04.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/datanode"
ssh $1@fa20-cs425-g13-04.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/namenode"
ssh $1@fa20-cs425-g13-05.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/datanode"
ssh $1@fa20-cs425-g13-05.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/namenode"
ssh $1@fa20-cs425-g13-06.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/datanode"
ssh $1@fa20-cs425-g13-06.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/namenode"
ssh $1@fa20-cs425-g13-07.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/datanode"
ssh $1@fa20-cs425-g13-07.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/namenode"
ssh $1@fa20-cs425-g13-08.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/datanode"
ssh $1@fa20-cs425-g13-08.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/namenode"
ssh $1@fa20-cs425-g13-09.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/datanode"
ssh $1@fa20-cs425-g13-09.cs.illinois.edu "mkdir -p /home/$1/hadoop_store/hdfs/namenode"

# copy hosts file
#cat /etc/hosts | ssh $1@fa20-cs425-g13-09.cs.illinois.edu -t "sudo tee -a /etc/hosts"

#hdfs namenode -format
#start-dfs.sh
#start-yarn.sh


#172.22.156.42 namenode
#172.22.158.42 datanode1
#172.22.94.42 datanode2
#172.22.156.43 datanode3
#172.22.158.43 datanode4
#172.22.94.43 datanode5
#172.22.156.44 datanode6
#172.22.158.44 datanode7
#172.22.94.44 datanode8
