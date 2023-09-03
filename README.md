## Dior Kafka/NSQ 压测工具

## 与Flume类似的三个概念

* Source
* Channel
* Sink

## Source Type :

* KafkaSource
* NSQSource
* PressSource : 压力测试源，指定一个源文件，数据之间行分割，作为压测数据，以指定速率向Sink发射数据

## Sink Type :

* KafkaSink
* NSQSink
* FileSink
* NilSink

## 编译环境：

* go 1.20
* GNU Make 3.81

## 使用方式Usage
### file to file

<pre>./build/dior -src press --src-file source.txt --src-speed 10 --dst file --dst-file sink.txt</pre>

### kafka to kafka

<pre>./build/dior -src kafka --src-bootstrap-servers./build/dior -src press --src-data-file data.txt --src-speed 1000 --dst kafka --dst-bootstrap-servers 127.0.0.1:9092 --dst-topic topic_to</pre>

### kafka to file

<pre></pre>

### nsq to file

<pre></pre>

### kafka to nsq

<pre></pre>

### nsq to kafka

<pre></pre>
