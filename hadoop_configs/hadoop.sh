# Copy ssh to other machines if namenode 
#ssh-copy-id -i /home/jit2/.ssh/id_rsa.pub jit2@fa20-cs425-g13-02.cs.illinois.edu
#ssh-copy-id -i /home/jit2/.ssh/id_rsa.pub jit2@fa20-cs425-g13-03.cs.illinois.edu
#ssh-copy-id -i /home/jit2/.ssh/id_rsa.pub jit2@fa20-cs425-g13-04.cs.illinois.edu
#ssh-copy-id -i /home/jit2/.ssh/id_rsa.pub jit2@fa20-cs425-g13-05.cs.illinois.edu
#ssh-copy-id -i /home/jit2/.ssh/id_rsa.pub jit2@fa20-cs425-g13-06.cs.illinois.edu
#ssh-copy-id -i /home/jit2/.ssh/id_rsa.pub jit2@fa20-cs425-g13-07.cs.illinois.edu
#ssh-copy-id -i /home/jit2/.ssh/id_rsa.pub jit2@fa20-cs425-g13-08.cs.illinois.edu
#ssh-copy-id -i /home/jit2/.ssh/id_rsa.pub jit2@fa20-cs425-g13-09.cs.illinois.edu

# Run this in /home/netid
# Download and install hadoop
#cd ~
#wget https://mirrors.gigenet.com/apache/hadoop/common/hadoop-2.9.2/hadoop-2.9.2.tar.gz
#tar -xzvf hadoop-2.9.2.tar.gz
#
#echo 'export HADOOP_HOME=/home/jit2/hadoop-2.9.2' >> ~/.bashrc
#echo 'export HADOOP_INSTALL=$HADOOP_HOME' >> ~/.bashrc
#echo 'export HADOOP_MAPRED_HOME=$HADOOP_HOME' >> ~/.bashrc
#echo 'export HADOOP_COMMON_HOME=$HADOOP_HOME' >> ~/.bashrc
#echo 'export HADOOP_HDFS_HOME=$HADOOP_HOME' >> ~/.bashrc
#echo 'export YARN_HOME=$HADOOP_HOME' >> ~/.bashrc
#echo 'export HADOOP_COMMON_LIB_NATIVE_DIR=$HADOOP_HOME/lib/native' >> ~/.bashrc
#echo 'export PATH=$PATH:$HADOOP_HOME/sbin:$HADOOP_HOME/bin' >> ~/.bashrc
#echo 'export HADOOP_OPTS="-Djava.library.path=$HADOOP_HOME/lib/native"' >> ~/.bashrc
#source ~/.bashrc
#
# Set up hadoop

#echo 'export JAVA_HOME=${JAVA_HOME:-"/usr/lib/jvm/java-1.8.0-openjdk-1.8.0.272.b10-1.el7_9.x86_64"}' >> /home/jit2/hadoop-2.9.2/etc/hadoop/hadoop-env.sh
#echo 'export HADOOP_CONF_DIR=${HADOOP_CONF_DIR:-"/home/jit2/hadoop-2.9.2/etc/hadoop"}' >> /home/jit2/hadoop-2.9.2/etc/hadoop/hadoop-env.sh
#
## Set up ips
#echo '172.22.156.42' >> /home/jit2/hadoop-2.9.2/etc/hadoop/masters
## echo '172.22.156.42' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves
#echo '172.22.158.42' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves #02
#echo '172.22.94.42' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves #03
#echo '172.22.156.43' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves #04
#echo '172.22.158.43' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves #05
#echo '172.22.94.43' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves #06
#echo '172.22.156.44' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves #07
#echo '172.22.158.44' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves #08
#echo '172.22.94.44' >> /home/jit2/hadoop-2.9.2/etc/hadoop/slaves #09
#
## Install hadoop on worker nodes
#su hadoop
#cd /home/jit2
#scp -r hadoop jit2@fa20-cs425-g13-02.cs.illinois.edu:/opt/hadoop
#scp -r hadoop jit2@fa20-cs425-g13-03.cs.illinois.edu:/opt/hadoop
#scp -r hadoop jit2@fa20-cs425-g13-04.cs.illinois.edu:/opt/hadoop
#scp -r hadoop jit2@fa20-cs425-g13-05.cs.illinois.edu:/opt/hadoop
#scp -r hadoop jit2@fa20-cs425-g13-06.cs.illinois.edu:/opt/hadoop
#scp -r hadoop jit2@fa20-cs425-g13-07.cs.illinois.edu:/opt/hadoop
#scp -r hadoop jit2@fa20-cs425-g13-08.cs.illinois.edu:/opt/hadoop
#scp -r hadoop jit2@fa20-cs425-g13-09.cs.illinois.edu:/opt/hadoop
#
