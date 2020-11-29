#!/bin/bash
# Run this in /home/netid
# Download and install hadoop
cd ~
wget https://www.apache.org/dyn/closer.cgi/hadoop/common/hadoop-2.9.2/hadoop-2.9.2.tar.gz
tar -xzvf hadoop-2.9.2.tar.gz

echo 'export HADOOP_HOME=/home/jit2/hadoop-2.9.2' >> ~/.bashrc
echo 'export HADOOP_INSTALL=$HADOOP_HOME' >> ~/.bashrc
echo 'export HADOOP_MAPRED_HOME=$HADOOP_HOME' >> ~/.bashrc
echo 'export HADOOP_COMMON_HOME=$HADOOP_HOME' >> ~/.bashrc
echo 'export HADOOP_HDFS_HOME=$HADOOP_HOME' >> ~/.bashrc
echo 'export YARN_HOME=$HADOOP_HOME' >> ~/.bashrc
echo 'export HADOOP_COMMON_LIB_NATIVE_DIR=$HADOOP_HOME/lib/native' >> ~/.bashrc
echo 'export PATH=$PATH:$HADOOP_HOME/sbin:$HADOOP_HOME/bin' >> ~/.bashrc
echo 'export HADOOP_OPTS="-Djava.library.path=$HADOOP_HOME/lib/native"' >> ~/.bashrc
source ~/.bashrc

# Set up hadoop

#echo 'export JAVA_HOME=${JAVA_HOME:-"/usr/lib/jvm/java-1.8.0-openjdk-1.8.0.272.b10-1.el7_9.x86_64"}' >> /home/jit2/hadoop-2.9.2/etc/hadoop/hadoop-env.sh
#echo 'export HADOOP_CONF_DIR=${HADOOP_CONF_DIR:-"/home/jit2/hadoop-2.9.2/etc/hadoop"}' >> /home/jit2/hadoop-2.9.2/etc/hadoop/hadoop-env.sh
