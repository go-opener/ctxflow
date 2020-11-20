create database demo;
use demo;
CREATE TABLE `demoUser` (
  `uid` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(512)  NOT NULL DEFAULT '' COMMENT '用户名',
  `age` int(20) unsigned NOT NULL DEFAULT '0' COMMENT '年龄',
  `last_modify_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8 COMMENT='用户demo';
insert into demoUser (`name`,`age`) values ("张三","20");