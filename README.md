# spider91

#### 介绍
91视频网站爬虫工具，可以批量或单独爬取视频。
不带参数运行程序时，进入日常爬取模式，固定每天3点爬取24小时内发布的30个评分最高的视频，评分由关键字、视频时长、作者分三项评分组成(score下的两个txt定义了关键词评分和作者评分，分数范围[-∞，100])。每周六4点会爬取本周评分最高的30个最热视频并把当周的视频整理到一个文件夹下。程序有去重机制不会重复下载同一个视频。

#### 软件架构
基于go1.15编写，依赖chrome浏览器、python下的m3_dl、pysocks。


#### 安装教程

1.  安装chrome浏览器。
2.  安装python、m3_dl、pysocks  
 pip3 install m3_dl  
 pip3 install pysocks  
3.  编译代码  
工程根目录下执行  
go mod tidy  
go build


#### 使用说明

1.    参数说明  
-c 爬取页面  
-u 爬取的网页 可以是单个视频的页面也可以使是类似首页的多个视频的页面。  
-o 视频存储路径  
-p 代理地址  
-t 同时爬取的视频个数  

2.    示例  
**单个视频爬取**  
   ./spider91 -c -u "http://91porn.com/view_video.php?viewkey=8cd0148b3fe08d4a4c2f" -p "http://127.0.0.1:10808"  
**单页多个视频爬取**  
   ./spider91 -c -u "http://91porn.com/v.php?category=rf&viewtype=basic&page=2" -p "http://127.0.0.1:10808"   
    






